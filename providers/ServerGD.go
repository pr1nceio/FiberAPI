package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cradio/gormx"
	fiberapi "github.com/fruitspace/FiberAPI"
	"github.com/fruitspace/FiberAPI/models/db"
	"github.com/fruitspace/FiberAPI/models/gdps_db"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/fruitspace/FiberAPI/services"
	"github.com/fruitspace/FiberAPI/utils"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//region ServerGDProvider

type ServerGDProvider struct {
	db       *gorm.DB
	mdb      *utils.MultiSQL
	redis    *utils.MultiRedis
	payments *PaymentProvider
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

func (sgp *ServerGDProvider) WithPaymentsProvider(pm *PaymentProvider) *ServerGDProvider {
	sgp.payments = pm
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

type ServerGD struct {
	srv        *db.ServerGd
	coreConfig *structs.GDPSConfig
	tariff     *structs.GDTariff
	p          *ServerGDProvider
}

//region Getters

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

//endregion

//region Settings

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

//endregion

//region Logs

func (s *ServerGD) GetLogs(xtype int, page int) ([]*gdps_db.Action, int, error) {

	qdb, err := s.p.mdb.OpenMutated("gdps", s.srv.SrvID)
	defer s.p.mdb.DisposeMutated("gdps", s.srv.SrvID)
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

	var results []*gdps_db.Action

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

//endregion

//region Songs

func (s *ServerGD) SearchSongs(query string, page int, mode string) ([]*gdps_db.Song, int, error) {
	a := gdps_db.Song{}
	mus := services.InitMusic(s.p.redis)

	qdb, err := s.p.mdb.OpenMutated("gdps", s.srv.SrvID)
	defer s.p.mdb.DisposeMutated("gdps", s.srv.SrvID)
	if utils.Should(err) != nil {
		log.Println(err)
		return nil, 0, err
	}

	qdb = s.p.mdb.UTable(qdb, a.TableName())

	var songs []*gdps_db.Song

	orderBy := "downloads DESC"
	if mode == "id" {
		orderBy = "id ASC"
	}
	if mode == "alpha" {
		orderBy = "name ASC"
	}

	if query != "" {
		qdb = qdb.Where("name LIKE ?", fmt.Sprintf("%%%s%%", query)).
			Or("artist LIKE ?", fmt.Sprintf("%%%s%%", query)).
			Or("id=?", query)
	}

	err = qdb.Order(orderBy).Limit(10).Offset(page * 10).Find(&songs).Error
	if utils.Should(err) != nil {
		log.Println(err)
		return nil, 0, err
	}

	for _, song := range songs {
		// Transform HAL resource
		if strings.HasPrefix(song.URL, "hal:") {
			xmus, err := mus.TransformHalResource(song.URL)
			if err != nil {
				continue
			}
			song.URL = xmus.Url
		}
		if strings.Contains(song.URL, "mediapool.halhost.cc") {
			song.URL = strings.ReplaceAll(song.URL, "mediapool.halhost.cc", "mus.fruitspace.one")
		}
	}

	return songs, s.getSongCount(qdb), nil
}

func (s *ServerGD) getSongCount(gdb *gorm.DB) int {
	var cnt int64
	gdb.Count(&cnt)
	return int(cnt)
}

func (s *ServerGD) AddSong(xtype string, url string) (*gdps_db.Song, error) {
	a := gdps_db.Song{}
	mus := services.InitMusic(s.p.redis)

	qdb, err := s.p.mdb.OpenMutated("gdps", s.srv.SrvID)
	defer s.p.mdb.DisposeMutated("gdps", s.srv.SrvID)
	if utils.Should(err) != nil {
		log.Println(err)
		return nil, err
	}

	qdb = s.p.mdb.UTable(qdb, a.TableName())

	var resp *structs.MusicResponse
	var rid string
	switch xtype {
	case "ng":
		rid, resp, err = mus.GetNG(url)
	case "yt":
		rid, resp, err = mus.GetYT(url)
	case "dz":
		rid, resp, err = mus.GetDZ(url)
	case "vk":
		rid, resp, err = mus.GetVK(url)
	case "db":
		resp = &structs.MusicResponse{
			Status: "ok",
			Name:   "Dropbox (Rename in DB)",
			Artist: "Dropbox",
			Size:   "5.00",
			Url:    url,
		}
	default:
		err = errors.New("unknown type")
	}
	if err != nil {
		return nil, err
	}
	id := s.PushSong(qdb, resp, xtype, rid)
	if id == 0 {
		return nil, errors.New("failed to upload song")
	}
	return &gdps_db.Song{
		ID:        id,
		Name:      resp.Name,
		AuthorID:  0,
		Artist:    resp.Artist,
		Size:      5.00,
		URL:       resp.Url,
		Downloads: 0,
		IsBanned:  false,
	}, nil
}

func (s *ServerGD) PushSong(qdb *gorm.DB, response *structs.MusicResponse, xtype string, rid string) int {
	if f, _ := regexp.MatchString(`[^0-9\.]`, response.Size.String()); f || response.Size == "" {
		response.Size = "5.00"
	}

	sz, _ := response.Size.Float64()
	song := gdps_db.Song{
		Name:   response.Name,
		Artist: response.Artist,
		Size:   sz,
		URL:    fmt.Sprintf("hal:%s:%s", xtype, rid),
	}
	if err := utils.Should(qdb.Create(&song).Error); err != nil {
		log.Println(err)
		return 0
	}
	return song.ID
}

//endregion

func (s *ServerGD) UpgradeServer(uid int, srvid string, tariffid int, duration string, promocode string) error {
	pm := NewPromocodeProvider(s.p.db)

	preg := regexp.MustCompile("^[a-zA-Z0-9]+$")
	if !preg.MatchString(srvid) || !s.Exists(srvid) {
		return errors.New("Invalid srvid |srvid")
	}
	s.GetServerBySrvID(srvid)
	if uid != s.srv.OwnerID && uid != 1 {
		return errors.New("Invalid owner")
	}

	if tariffid < s.srv.Plan || tariffid > len(fiberapi.ProductGDTariffs) {
		return errors.New("Invalid tariff |tariff")
	}
	tariff := fiberapi.ProductGDTariffs[strconv.Itoa(tariffid)]

	when := time.Now()
	// Select which is latter (only non-free tariffs)
	if when.Compare(s.srv.ExpireDate) > 0 && s.srv.Plan > 1 {
		when = s.srv.ExpireDate
	}
	if when.Year() > 2040 && duration != "all" {
		return errors.New("Invalid duration |dur")
	}
	switch duration {
	case "all":
		when = time.Date(2050, 1, 2, 0, 0, 0, 0, time.UTC)
	case "yr":
		when = when.AddDate(1, 0, 0)
	default:
		when = when.AddDate(0, 1, 0)
	}

	if tariff.PriceRUB != 0 {
		price := float64(tariff.PriceRUB)
		if duration == "yr" {
			price *= 10
		}
		if duration == "all" {
			price *= 30 // 3*10
		}

		if promocode != "" {
			promo := pm.Get(promocode)
			if promo == nil {
				return errors.New("Invalid promocode |promo_invalid")
			}
			prc, err := promo.Use(price, "gd", strconv.Itoa(tariffid))
			if err != nil {
				return err
			}
			price = prc
		}

		price--

		resp := s.p.payments.SpendMoney(uid, price)
		if resp.Status != "ok" {
			return errors.New(resp.Message)
		}
	}

	if err := s.p.db.Model(&s.srv).WhereBinary(db.ServerGd{SrvID: srvid}).Updates(db.ServerGd{ExpireDate: when, Plan: tariffid}); err != nil {
		log.Println(err)
		return errors.New("DATABASE ERROR. REPORT IMMEDIATELY")
	}
	s.coreConfig.ServerConfig.MaxUsers = tariff.Players
	s.coreConfig.ServerConfig.MaxLevels = tariff.Levels
	s.coreConfig.ServerConfig.MaxPosts = tariff.Posts
	s.coreConfig.ServerConfig.MaxComments = tariff.Comments
	s.coreConfig.ServerConfig.Locked = false

	vdata, _ := json.Marshal(s.coreConfig)
	return utils.Should(s.p.redis.Get("gdps").Set(context.Background(), srvid, string(vdata), 0).Err())

}

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
