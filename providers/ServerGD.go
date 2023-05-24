package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cradio/gormx"
	"github.com/fruitspace/FiberAPI/models/db"
	"github.com/fruitspace/FiberAPI/models/gdps_db"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/fruitspace/FiberAPI/utils"
	"log"
	"strconv"
	"strings"
)

//region ServerGDProvider

type ServerGDProvider struct {
	db       *gorm.DB
	mdb      *utils.MultiSQL
	redis    *utils.MultiRedis
	keys     map[string]string
	config   map[string]string
	s3config map[string]string
}

func NewServerGDProvider(db *gorm.DB, mdb *utils.MultiSQL, redis *utils.MultiRedis) *ServerGDProvider {
	mdb.AddMutator("gdps", func(db string) string {
		return "gdps_" + db
	})
	return &ServerGDProvider{db: db, mdb: mdb, redis: redis}
}

func (sgp *ServerGDProvider) WithKeys(keys, config, s3config map[string]string) *ServerGDProvider {
	sgp.keys = keys
	sgp.config = config
	sgp.s3config = s3config
	return sgp
}

func (sgp *ServerGDProvider) New() *ServerGD {
	return &ServerGD{srv: &db.ServerGd{}, p: sgp}
}

func (sgp *ServerGDProvider) GetUserServers(uid int) []*db.ServerGdSmall {
	var srvs []*db.ServerGdSmall
	sgp.db.Model(db.ServerGd{}).Where(db.ServerGd{OwnerID: uid}).Find(&srvs)
	for _, srv := range srvs {
		srv.Icon = "https://" + sgp.s3config["cdn"] + "/server_icons/" + srv.Icon
	}

	return srvs
}

//endregion

//region ServerGD

type ServerGD struct {
	srv        *db.ServerGd
	coreConfig *structs.GDPSConfig
	tariff     *structs.GDTariff
	p          *ServerGDProvider
}

func (s *ServerGD) GetReducedServer(srvid string) (srv db.ServerGdReduced) {
	// Empty db.User for convenience
	u := db.User{}
	s.p.db.Model(db.ServerGd{}).SelectFields(srv,
		fmt.Sprintf("(SELECT %s FROM %s WHERE %s=%s) as owner",
			gorm.Column(u, "Uname"),
			u.TableName(),
			gorm.Column(u, "UID"),
			gorm.Column(db.ServerGd{}, "OwnerID"),
			// Where BINARY SrvID = srvid
		)).WhereBinary(db.ServerGd{SrvID: srvid}).Find(&srv)
	srv.Icon = "https://" + s.p.s3config["cdn"] + "/server_icons/" + srv.Icon
	srv.Description = strings.ReplaceAll(srv.Description, "#levels#", strconv.Itoa(srv.LevelCount))
	srv.Description = strings.ReplaceAll(srv.Description, "#players#", strconv.Itoa(srv.UserCount))
	return srv
}

func (s *ServerGD) GetTopUserServer(uid int) (srv db.ServerGdSmall) {
	s.p.db.Model(db.ServerGd{}).Where(db.ServerGd{OwnerID: uid}).Find(&srv)
	srv.Icon = "https://" + s.p.s3config["cdn"] + "/server_icons/" + srv.Icon
	return srv
}

func (s *ServerGD) Exists(srvid string) bool {
	var cnt int64
	s.p.db.Model(s.srv).WhereBinary(db.ServerGd{SrvID: srvid}).Count(&cnt)
	return cnt > 0
}

func (s *ServerGD) GetServerBySrvID(srvid string) bool {
	return s.p.db.WhereBinary(db.ServerGd{SrvID: srvid}).First(&s.srv).Error == nil
}

func (s *ServerGD) LoadCoreConfig() (err error) {
	if v, err := s.p.redis.Get("gdps").Get(context.Background(), s.srv.SrvID).Result(); err == nil {
		return json.Unmarshal([]byte(v), &s.coreConfig)
	}
	return err
}

func (s *ServerGD) GetTextures() string {
	if s.srv.IsCustomTextures {
		return s.srv.SrvID + ".zip"
	} else {
		return "gdps_textures.zip"
	}
}

func (s *ServerGD) ResetDBPassword() error {

	pwd := utils.GenString(12) + "*"
	s.coreConfig.DBConfig.Password = pwd
	s.srv.DbPassword = pwd

	tx, err := s.p.mdb.Raw().BeginTx(context.Background(), nil)
	tx.Exec(fmt.Sprintf("ALTER USER halgd_%s@localhost IDENTIFIED BY ?", s.srv.SrvID), pwd)
	tx.Exec(fmt.Sprintf("ALTER USER halgd_%s@'%%' IDENTIFIED BY ?", s.srv.SrvID), pwd)
	err = tx.Commit()
	if utils.Should(err) != nil {
		log.Println(err)
	} else {
		updated, err := json.Marshal(s.coreConfig)
		err = s.p.redis.Get("gdps").Set(context.Background(), s.srv.SrvID, string(updated), 0).Err()
		if utils.Should(err) != nil {
			log.Println(err)
		} else {
			err = s.p.db.Model(&s.srv).WhereBinary(db.ServerGd{SrvID: s.srv.SrvID}).Updates(db.ServerGd{DbPassword: pwd}).Error
		}
	}
	return err
}

func (s *ServerGD) UpdateSettings(settings structs.GDSettings) error {
	s.srv.Description = settings.Description.Text
	s.srv.TextAlign = settings.Description.Align
	ds := strings.Split(settings.Description.Discord, "/")
	s.srv.Discord = ds[len(ds)-1]
	vk := strings.Split(settings.Description.Vk, "/")
	s.srv.Vk = vk[len(vk)-1]

	if s.srv.IsSpaceMusic == false {
		s.srv.IsSpaceMusic = settings.SpaceMusic
		//If enabled -> update core config
		if s.srv.IsSpaceMusic == true {
			s.coreConfig.ServerConfig.HalMusic = true
			updated, err := json.Marshal(s.coreConfig)
			err = s.p.redis.Get("gdps").Set(context.Background(), s.srv.SrvID, string(updated), 0).Err()
			if utils.Should(err) != nil {
				log.Println(err)
				return err
			}
		}
	}

	return s.p.db.Model(&s.srv).WhereBinary(db.ServerGd{SrvID: s.srv.SrvID}).Updates(db.ServerGd{
		Description:  s.srv.Description,
		TextAlign:    s.srv.TextAlign,
		Discord:      s.srv.Discord,
		Vk:           s.srv.Vk,
		IsSpaceMusic: s.srv.IsSpaceMusic,
	}).Error
}

func (s *ServerGD) UpdateChests(chests structs.ChestConfig) error {
	s.coreConfig.ChestConfig = chests
	updated, err := json.Marshal(chests)

	err = s.p.redis.Get("gdps").Set(context.Background(), s.srv.SrvID, string(updated), 0).Err()
	if utils.Should(err) != nil {
		log.Println(err)
	}
	return err
}

func (s *ServerGD) GetLogs(xtype int, page int) ([]gdps_db.Action, int, error) {

	qdb, err := s.p.mdb.OpenMutated("gdps", s.srv.SrvID)
	if utils.Should(err) != nil {
		log.Println(err)
		return nil, 0, err
	}
	a := gdps_db.Action{}

	qdb = s.p.mdb.UTable(qdb, a.TableName())
	pqdb := qdb // For count
	rqdb := qdb.Select(utils.HideField(a, "Data"), fmt.Sprintf(`
	JSON_INSERT(
		JSON_INSERT(actions.data,
		    '$.name',
			CASE WHEN actions.type=4 THEN
				(SELECT name FROM levels WHERE levels.id=actions.target_id)
		    ELSE
		    	NULL
		    END
		),
		'$.uname',
		CASE WHEN actions.type=4 THEN
			(SELECT uname FROM users WHERE users.uid=actions.uid)
		ELSE
			NULL
		END
	) as data`)).Limit(50).Offset(page * 50)

	var results []gdps_db.Action

	if xtype >= 0 {
		rqdb = rqdb.Where(gdps_db.Action{Type: xtype})
		pqdb = pqdb.Where(gdps_db.Action{Type: xtype})
	} else {
		rqdb = rqdb.Where(fmt.Sprintf("%s<6", gorm.Column(a, "Type")))
		pqdb = pqdb.Where(fmt.Sprintf("%s<6", gorm.Column(a, "Type")))
	}

	if err = rqdb.Find(&results).Error; utils.Should(err) != nil {
		log.Println(err)
		return nil, 0, err
	}

	var cnt int64
	err = pqdb.Count(&cnt).Error
	if cnt%50 == 0 {
		cnt = cnt / 50
	} else {
		cnt = cnt/50 + 1
	}

	return results, int(cnt), err
}

//func (s *ServerGD) SearchSongs(query string, page int, mode string) ([]byte, error) {
//	pack, err := json.Marshal(struct {
//		Page  int    `json:"page"`
//		Query string `json:"query"`
//		Mode  string `json:"mode"`
//	}{page, query, mode})
//	r, err := http.Post(SCHEDULER_HOST+"/gd/"+srv.SrvId+"/get_music", "text/json", bytes.NewReader(pack))
//	if err != nil {
//		return nil, err
//	}
//	resp, _ := io.ReadAll(r.Body)
//	return resp, err
//}

//func (srv *GDServer) AddSong(xtype string, url string) ([]byte, error) {
//	pack, err := json.Marshal(struct {
//		Type string `json:"type"`
//		Url  string `json:"url"`
//	}{xtype, url})
//	r, err := http.Post(SCHEDULER_HOST+"/gd/"+srv.SrvId+"/add_music", "text/json", bytes.NewReader(pack))
//	if err != nil {
//		return nil, err
//	}
//	resp, _ := io.ReadAll(r.Body)
//	return resp, err
//}
//
//func (srv *GDServer) UpgradeServer(uid int64, srvid string, tariffid int, duration string, promocode string) error {
//	preg := regexp.MustCompile("^[a-zA-Z0-9]+$")
//	if !preg.MatchString(srvid) || !srv.Exists(srvid) {
//		return errors.New("Invalid srvid |srvid")
//	}
//	srv.SrvId = srvid
//	srv.LoadAll()
//	if uid != srv.OwnerId && uid != 1 {
//		return errors.New("Invalid owner")
//	}
//
//	if tariffid < srv.Plan || tariffid > len(Structures.ProductGDTariffs) {
//		return errors.New("Invalid tariff |tariff")
//	}
//	tariff := Structures.ProductGDTariffs[strconv.Itoa(tariffid)]
//
//	when, err := time.Parse("2006-01-02 15:04:05", srv.ExpireDate)
//	if err != nil {
//		return err
//	}
//	if srv.Plan < 2 {
//		when = time.Now()
//	}
//	if when.Year() > 2040 && duration != "all" {
//		return errors.New("Invalid duration |dur")
//	}
//	switch duration {
//	case "all":
//		when = time.Date(2050, 1, 2, 0, 0, 0, 0, time.UTC)
//	case "yr":
//		when = when.AddDate(1, 0, 0)
//	default:
//		when = when.AddDate(0, 1, 0)
//	}
//	whenText := when.Format("2006/01/02")
//	req := struct {
//		Players  int  `json:"players"`
//		Levels   int  `json:"levels"`
//		Posts    int  `json:"posts"`
//		Comments int  `json:"comments"`
//		Locked   bool `json:"locked"`
//	}{tariff.Players, tariff.Levels, tariff.Posts, tariff.Comments, false}
//	pack, _ := json.Marshal(req)
//
//	if tariff.PriceRUB != 0 {
//		price := float64(tariff.PriceRUB)
//		if duration == "yr" {
//			price *= 10
//		}
//		if duration == "all" {
//			price *= 30 // 3*10
//		}
//
//		if promocode != "" {
//			promo := GetPromocode(promocode)
//			if promo.Id == 0 {
//				return errors.New("Invalid promocode |promo_invalid")
//			}
//			prc, err := promo.Use(price, "gd", strconv.Itoa(tariffid))
//			if err != nil {
//				return err
//			}
//			price = prc
//		}
//
//		price--
//
//		resp := CTransaction{}.SpendMoney(uid, price)
//		if resp.Status != "ok" {
//			return errors.New(resp.Message)
//		}
//	}
//
//	_, err = DB.Exec("UPDATE servers_gd SET expireDate=?,plan=? WHERE BINARY srvid=?", whenText, tariffid, srvid)
//	if err != nil {
//		log.Println(err.Error())
//		return errors.New("DATABASE ERROR. REPORT IMMEDIATELY")
//	}
//
//	r, err := http.Post(SCHEDULER_HOST+"/gd/"+srvid+"/set_limits", "text/json", bytes.NewReader(pack))
//	if err != nil {
//		return err
//	}
//	resp := struct {
//		Status string `json:"status"`
//	}{}
//	json.NewDecoder(r.Body).Decode(&resp)
//	if resp.Status != "ok" {
//		return errors.New("Internal creation error |internal")
//	}
//	return nil
//
//}
//
//func (srv *GDServer) CreateServer(uid int64, name string, tariffid int, duration string, promocode string) error {
//	preg := regexp.MustCompile("^[a-zA-Z0-9 ._-]+$")
//	if !preg.MatchString(name) {
//		return errors.New("Invalid name |name")
//	}
//	if tariffid < 1 || tariffid > len(Structures.ProductGDTariffs) {
//		return errors.New("Invalid tariff |tariff")
//	}
//	tariff := Structures.ProductGDTariffs[strconv.Itoa(tariffid)]
//	when := time.Now()
//	switch duration {
//	case "all":
//		when = time.Date(2050, 1, 2, 0, 0, 0, 0, time.UTC)
//	case "yr":
//		when = when.AddDate(1, 0, 0)
//	default:
//		when = when.AddDate(0, 1, 0)
//	}
//
//	whenText := when.Format("2006/01/02")
//
//	req := struct {
//		Uid       int64  `json:"uid"`
//		Plan      int    `json:"plan"`
//		Name      string `json:"name"`
//		Expire    string `json:"expire"`
//		Is22      bool   `json:"is22"`
//		MaxUsers  int    `json:"maxUsers"`
//		MaxLevels int    `json:"maxLevels"`
//	}{uid, tariffid, name, whenText, false, tariff.Players, tariff.Levels}
//	pack, _ := json.Marshal(req)
//
//	if tariffid == 1 {
//		var cnt int
//		DB.Get(&cnt, "SELECT COUNT(*) FROM servers_gd WHERE owner_id=? AND plan=1", uid)
//		if cnt != 0 {
//			return errors.New("You already have FREE server")
//		}
//	}
//
//	if tariff.PriceRUB != 0 {
//		price := float64(tariff.PriceRUB)
//		if duration == "yr" {
//			price *= 10
//		}
//		if duration == "all" {
//			price *= 30 // 3*10
//		}
//
//		if promocode != "" {
//			promo := GetPromocode(promocode)
//			if promo.Id == 0 {
//				return errors.New("Invalid promocode |promo_invalid")
//			}
//			prc, err := promo.Use(price, "gd", strconv.Itoa(tariffid))
//			if err != nil {
//				return err
//			}
//			price = prc
//		}
//
//		price--
//
//		resp := CTransaction{}.SpendMoney(uid, price)
//		if resp.Status != "ok" {
//			return errors.New(resp.Message)
//		}
//	}
//
//	r, err := http.Post(SCHEDULER_HOST+"/gd/new", "text/json", bytes.NewReader(pack))
//	if err != nil {
//		return err
//	}
//	resp := struct {
//		Status string `json:"status"`
//		SrvId  string `json:"srvId"`
//	}{}
//	json.NewDecoder(r.Body).Decode(&resp)
//	if resp.Status != "ok" || len(resp.SrvId) != 4 {
//		return errors.New("Internal creation error |internal")
//	}
//	return nil
//}
//
//func (srv *GDServer) UpdateLogo(img []byte) error {
//	theImg, _, _ := image.Decode(bytes.NewReader(img))
//	cropImg := lib.CropSquareImage(theImg)
//	var buf bytes.Buffer
//	if err := Logger.Should(png.Encode(&buf, cropImg)); err != nil {
//		return err
//	}
//	newImg, _ := io.ReadAll(&buf)
//
//	creds := credentials.NewStaticCredentials(S3_CONFIG["access_key"], S3_CONFIG["secret"], "")
//	cfg := aws.NewConfig().WithEndpoint(S3_CONFIG["endpoint"]).WithRegion(S3_CONFIG["region"]).WithCredentials(creds)
//	sess, err := session.NewSession()
//	if Logger.Should(err) != nil {
//		return err
//	}
//	svc := s3.New(sess, cfg)
//	srv.Icon = "gd_" + srv.SrvId + ".png"
//
//	params := &s3.PutObjectInput{
//		Bucket:        aws.String(S3_CONFIG["bucket"]),
//		Key:           aws.String("server_icons/" + srv.Icon),
//		Body:          bytes.NewReader(newImg),
//		ContentLength: aws.Int64(int64(len(newImg))),
//		ContentType:   aws.String("image/png"),
//	}
//	_, err = svc.PutObject(params)
//	if Logger.Should(err) != nil {
//		return err
//	}
//	_, err = DB.Exec("UPDATE servers_gd SET icon=? WHERE id=?", srv.Icon, srv.Id)
//	return err
//}
//
//func (srv *GDServer) UploadTextures(inp io.Reader) error {
//
//	buf := bytes.NewBuffer(nil)
//	go func() {
//		_, _ = io.Copy(buf, inp)
//	}()
//
//	creds := credentials.NewStaticCredentials(S3_CONFIG["access_key"], S3_CONFIG["secret"], "")
//	cfg := aws.NewConfig().WithEndpoint(S3_CONFIG["endpoint"]).WithRegion(S3_CONFIG["region"]).WithCredentials(creds)
//	sess, err := session.NewSession()
//	if Logger.Should(err) != nil {
//		return err
//	}
//	svc := s3.New(sess, cfg)
//	srv.Icon = "gd_" + srv.SrvId + ".png"
//
//	params := &s3.PutObjectInput{
//		Bucket: aws.String(S3_CONFIG["bucket"]),
//		Key:    aws.String("server_icons/" + srv.Icon),
//		Body:   bytes.NewReader(buf.Bytes()),
//		//ContentLength: aws.Int64(int64(len(newImg))),
//		ContentType: aws.String("application/zip"),
//	}
//	_, err = svc.PutObject(params)
//	if Logger.Should(err) != nil {
//		return err
//	}
//
//	return nil
//}
//
//func (srv *GDServer) ExecuteBuildLab(conf BuildLabSettings) error {
//	if conf.SrvName != "" {
//		preg := regexp.MustCompile("^[a-zA-Z0-9 ._-]+$")
//		if !preg.MatchString(conf.SrvName) {
//			return errors.New("Invalid name |name")
//		}
//		srv.SrvName = conf.SrvName
//	} else {
//		conf.SrvName = srv.SrvName
//	}
//	if conf.Version != "2.2" {
//		conf.Version = "2.1"
//	}
//	if conf.Icon != "custom" {
//		conf.Icon = "gd_default.png"
//	}
//	//!Ignore textures
//	_, err := DB.Exec("UPDATE servers_gd SET srvName=? WHERE id=?", srv.SrvName, srv.Id)
//	if err != nil {
//		return err
//	}
//	pack, _ := json.Marshal(conf)
//	r, err := http.Post(SCHEDULER_HOST+"/gd/"+srv.SrvId+"/build", "text/json", bytes.NewReader(pack))
//	if err != nil {
//		return err
//	}
//	resp := struct {
//		Status string `json:"status"`
//	}{}
//	json.NewDecoder(r.Body).Decode(&resp)
//	if resp.Status != "ok" {
//		return errors.New("Internal build error |internal")
//	}
//	return nil
//}
//
//func (srv *GDServer) DeleteServer() error {
//	resp := make(map[string]string)
//	r, err := http.Get(SCHEDULER_HOST + "/gd/" + srv.SrvId + "/delete")
//	err = json.NewDecoder(r.Body).Decode(&resp)
//	if err != nil {
//		return err
//	}
//	if resp["status"] != "ok" {
//		return errors.New(resp["error"])
//	}
//	return err
//}

//endregion
