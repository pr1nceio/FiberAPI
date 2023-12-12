package providers

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cradio/gormx"
	fiberapi "github.com/fruitspace/FiberAPI"
	"github.com/fruitspace/FiberAPI/models/db"
	"github.com/fruitspace/FiberAPI/models/gdps_db"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/fruitspace/FiberAPI/services"
	"github.com/fruitspace/FiberAPI/utils"
	"image"
	"image/png"
	"io"
	"log"
	"net/http"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

var Jobs []string

//region ServerGDProvider

type ServerGDProvider struct {
	db          *gorm.DB
	mdb         *utils.MultiSQL
	redis       *utils.MultiRedis
	payments    *PaymentProvider
	assets      *embed.FS
	keys        map[string]string
	config      map[string]string
	s3config    map[string]string
	minioconfig map[string]string
}

func NewServerGDProvider(db *gorm.DB, mdb *utils.MultiSQL, redis *utils.MultiRedis) *ServerGDProvider {
	mdb.AddMutator("gdps", func(db string) string {
		return "gdps_" + db
	})
	return &ServerGDProvider{db: db, mdb: mdb, redis: redis}
}

func (sgp *ServerGDProvider) WithKeys(keys, config, s3config, minioconfig map[string]string) *ServerGDProvider {
	sgp.keys = keys
	sgp.config = config
	sgp.s3config = s3config
	sgp.minioconfig = minioconfig
	return sgp
}

func (sgp *ServerGDProvider) WithPaymentsProvider(pm *PaymentProvider) *ServerGDProvider {
	sgp.payments = pm
	return sgp
}

func (sgp *ServerGDProvider) WithAssets(assets *embed.FS) *ServerGDProvider {
	sgp.assets = assets
	return sgp
}

func (sgp *ServerGDProvider) New() *ServerGD {
	return &ServerGD{Srv: &db.ServerGd{}, p: sgp}
}

func (sgp *ServerGDProvider) ExposeRedis() *utils.MultiRedis {
	return sgp.redis
}

func (sgp *ServerGDProvider) ExposeGorm() *gorm.DB {
	return sgp.db
}

func (sgp *ServerGDProvider) GetUserServers(uid int) []*db.ServerGdSmall {
	var srvs []*db.ServerGdSmall
	sgp.db.Model(db.ServerGd{}).Where(db.ServerGd{OwnerID: uid}).Find(&srvs)
	for _, srv := range srvs {
		srv.Icon = "https://" + sgp.s3config["cdn"] + "/server_icons/" + srv.Icon
	}

	return srvs
}

func (sgp *ServerGDProvider) GetTopServers(offset int) []*db.ServerGdSmall {
	var srvs []*db.ServerGdSmall
	sgp.db.Model(db.ServerGd{}).Where(fmt.Sprintf("%s>1", gorm.Column(db.ServerGd{}, "Plan"))).
		Where(fmt.Sprintf("%s>NOW()", gorm.Column(db.ServerGd{}, "ExpireDate"))).
		Order(fmt.Sprintf("%s DESC", gorm.Column(db.ServerGd{}, "UserCount"))).
		Limit(10).Offset(offset).Find(&srvs)
	for _, srv := range srvs {
		srv.Icon = "https://" + sgp.s3config["cdn"] + "/server_icons/" + srv.Icon
		srv.ExpireDate = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
		srv.Plan = 0
	}

	return srvs
}

func (sgp *ServerGDProvider) CountServers() int {
	var cnt int64
	sgp.db.Model(db.ServerGd{}).Count(&cnt)
	return int(cnt)
}

func (sgp *ServerGDProvider) CountLevels() int {
	var cnt int64
	sgp.db.Table((&db.ServerGd{}).TableName()).Select(fmt.Sprintf("sum(%s)", gorm.Column(db.ServerGd{}, "LevelCount"))).Row().Scan(&cnt)
	return int(cnt)
}

func (sgp *ServerGDProvider) GetUnpaidServers() []string {
	var srvs []*db.ServerGdSmall
	var srvids []string
	sgp.db.Model(db.ServerGd{}).Where(fmt.Sprintf("%s<NOW()", gorm.Column(db.ServerGd{}, "ExpireDate"))).Find(&srvs)
	for _, srv := range srvs {
		srvids = append(srvids, srv.SrvID)
	}
	return srvids
}

func (sgp *ServerGDProvider) GetInactiveServers(maxUsers int, free bool) []string {
	var srvs []*db.ServerGdSmall
	var srvids []string
	tx := sgp.db.Model(db.ServerGd{}).
		Where(fmt.Sprintf("%s<(CURRENT_DATE - INTERVAL 7 DAY)", gorm.Column(db.ServerGd{}, "CreatedAt"))).
		Where(fmt.Sprintf("%s<=%d", gorm.Column(db.ServerGd{}, "UserCount"), maxUsers))
	if free {
		tx = tx.Where(db.ServerGd{Plan: 1})
	}
	tx.Find(&srvs)
	for _, srv := range srvs {
		srvids = append(srvids, srv.SrvID)
	}
	return srvids
}

func (sgp *ServerGDProvider) GetMissingInstallersServers() []string {
	var srvs []*db.ServerGdSmall
	var srvids []string
	tx := sgp.db.Model(db.ServerGd{}).
		//Where(fmt.Sprintf("%s<(CURRENT_DATE - INTERVAL 1 DAY)", gorm.Column(db.ServerGd{}, "CreatedAt"))).
		Where(fmt.Sprintf("%s=''", gorm.Column(db.ServerGd{}, "ClientWindowsURL")))
	tx.Find(&srvs)
	for _, srv := range srvs {
		srvids = append(srvids, srv.SrvID)
	}
	return srvids
}

//endregion

type ServerGD struct {
	Srv        *db.ServerGd
	CoreConfig *structs.GDPSConfig
	Tariff     *structs.GDTariff
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
	s.p.db.Model(db.ServerGd{}).Where(db.ServerGd{OwnerID: uid}).Order(fmt.Sprintf("%s DESC", gorm.Column(db.ServerGd{}, "UserCount"))).Find(&srv)
	srv.Icon = "https://" + s.p.s3config["cdn"] + "/server_icons/" + srv.Icon
	return srv
}

func (s *ServerGD) Exists(srvid string) bool {
	var cnt int64
	s.p.db.Model(s.Srv).WhereBinary(db.ServerGd{SrvID: srvid}).Count(&cnt)
	if cnt > 0 {
		s.Srv.SrvID = srvid
	}
	return cnt > 0
}

func (s *ServerGD) GetServerBySrvID(srvid string) bool {
	return s.p.db.WhereBinary(db.ServerGd{SrvID: srvid}).First(&s.Srv).Error == nil
}

func (s *ServerGD) LoadCoreConfig() (err error) {
	var v string
	if v, err = s.p.redis.Get("gdps").Get(context.Background(), s.Srv.SrvID).Result(); err == nil {
		return json.Unmarshal([]byte(v), &s.CoreConfig)
	}
	return err
}

func (s *ServerGD) LoadTariff() {
	t := fiberapi.ProductGDTariffs[strconv.Itoa(s.Srv.Plan)]
	s.Tariff = &t
}

func (s *ServerGD) GetTextures() string {
	if s.Srv.IsCustomTextures {
		return s.Srv.SrvID + ".zip"
	} else {
		return "gdps_textures.zip"
	}
}

//endregion

//region Settings

func (s *ServerGD) ResetDBPassword() error {
	s.LoadCoreConfig()

	pwd := utils.GenString(12) + "*"
	s.CoreConfig.DBConfig.Password = pwd
	s.Srv.DbPassword = pwd

	rawdb := s.p.mdb.Raw()
	_, err := rawdb.Exec(fmt.Sprintf("ALTER USER halgd_%s@localhost IDENTIFIED BY '%s'", s.Srv.SrvID, pwd))
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = rawdb.Exec(fmt.Sprintf("ALTER USER halgd_%s@'%%' IDENTIFIED BY '%s'", s.Srv.SrvID, pwd))
	if utils.Should(err) != nil {
		log.Println(err)
		return err
	}
	updated, _ := json.Marshal(s.CoreConfig)
	err = s.p.redis.Get("gdps").Set(context.Background(), s.Srv.SrvID, string(updated), 0).Err()
	if utils.Should(err) != nil {
		log.Println(err)
		return err
	}
	err = s.p.db.Model(&s.Srv).WhereBinary(db.ServerGd{SrvID: s.Srv.SrvID}).Updates(db.ServerGd{DbPassword: pwd}).Error
	return err
}

func (s *ServerGD) UpdateSettings(settings structs.GDSettings) error {
	defer func() {
		if l := recover(); l != nil {
			log.Println(l)
			m := fmt.Sprintf("%+v\n%+v", settings, s.CoreConfig)
			utils.SendMessageDiscord(m)
			log.Println(m)
		}
	}()

	if err := s.LoadCoreConfig(); err != nil {
		return err
	}
	s.Srv.Description = settings.Description.Text
	s.Srv.TextAlign = settings.Description.Align
	ds := strings.Split(settings.Description.Discord, "/")
	s.Srv.Discord = ds[len(ds)-1]
	vk := strings.Split(settings.Description.Vk, "/")
	s.Srv.Vk = vk[len(vk)-1]

	s.CoreConfig.SecurityConfig.DisableProtection = !settings.Security.Enabled
	s.CoreConfig.SecurityConfig.AutoActivate = settings.Security.AutoActivate
	s.CoreConfig.SecurityConfig.NoLevelLimits = !settings.Security.LevelLimit

	s.CoreConfig.ServerConfig.TopSize = settings.TopSize
	s.CoreConfig.ServerConfig.HalMusic = settings.SpaceMusic
	s.CoreConfig.ServerConfig.EnableModules = settings.Modules

	if s.Srv.IsSpaceMusic == false {
		s.Srv.IsSpaceMusic = settings.SpaceMusic
		//If enabled -> update core config
		if s.Srv.IsSpaceMusic == true {
			s.CoreConfig.ServerConfig.HalMusic = true
		}
	}

	updated, _ := json.Marshal(s.CoreConfig)
	err := s.p.redis.Get("gdps").Set(context.Background(), s.Srv.SrvID, string(updated), 0).Err()
	if utils.Should(err) != nil {
		log.Println(err)
		return err
	}

	return s.p.db.Model(&s.Srv).WhereBinary(db.ServerGd{SrvID: s.Srv.SrvID}).Updates(db.ServerGd{
		Description:  s.Srv.Description,
		TextAlign:    s.Srv.TextAlign,
		Discord:      s.Srv.Discord,
		Vk:           s.Srv.Vk,
		IsSpaceMusic: s.Srv.IsSpaceMusic,
	}).Error
}

func (s *ServerGD) UpdateChests(chests structs.ChestConfig) error {
	if errm := s.LoadCoreConfig(); errm != nil {
		return utils.Should(errm)
	}
	s.CoreConfig.ChestConfig = chests
	updated, _ := json.Marshal(s.CoreConfig)

	err := s.p.redis.Get("gdps").Set(context.Background(), s.Srv.SrvID, string(updated), 0).Err()
	if utils.Should(err) != nil {
		log.Println(err)
	}
	return err
}

//endregion

//region Logs

func (s *ServerGD) GetLogs(xtype int, page int) ([]*gdps_db.Action, int, error) {

	qdb, err := s.p.mdb.OpenMutated("gdps", s.Srv.SrvID)
	defer s.p.mdb.DisposeMutated("gdps", s.Srv.SrvID)
	if utils.Should(err) != nil {
		log.Println(err)
		return nil, 0, err
	}
	a := gdps_db.Action{}

	qdb = s.p.mdb.UTable(qdb, a.TableName())

	d := s.p.mdb.Mutate("gdps", s.Srv.SrvID)
	rqdb := qdb.Select(utils.HideField(a, "Data"), fmt.Sprintf(`
	JSON_INSERT(
		JSON_INSERT(%s.actions.data,
		    '$.name',
			CASE WHEN %s.actions.type=4 THEN
				(SELECT name FROM %s.levels WHERE %s.levels.id=%s.actions.target_id)
		    ELSE
		    	NULL
		    END
		),
		'$.uname',
		CASE WHEN %s.actions.type=4 THEN
			(SELECT uname FROM %s.users WHERE %s.users.uid=%s.actions.uid)
		ELSE
			NULL
		END
	) as data`, d, d, d, d, d, d, d, d, d,
	))

	var results []*gdps_db.Action

	if xtype >= 0 {
		rqdb = rqdb.Where(fmt.Sprintf("%s=?", gorm.Column(a, "Type")), xtype)
	} else {
		rqdb = rqdb.Where(fmt.Sprintf("%s<6", gorm.Column(a, "Type")))
	}

	var cnt int64
	rqdb.Count(&cnt)
	if cnt%50 == 0 {
		cnt = cnt / 50
	} else {
		cnt = cnt/50 + 1
	}

	rqdb = rqdb.Limit(50).Offset(page * 50)

	if err = rqdb.Find(&results).Error; utils.Should(err) != nil {
		log.Println(err)
		return nil, 0, err
	}

	return results, int(cnt), err
}

//endregion

//region Songs

//func (s *ServerGD) transfromSongs() error {
//	a := gdps_db.Song{}
//	mus := services.InitMusic(s.p.redis)
//
//	qdb, err := s.p.mdb.OpenMutated("gdps", s.Srv.SrvID)
//	defer s.p.mdb.DisposeMutated("gdps", s.Srv.SrvID)
//	if utils.Should(err) != nil {
//		log.Println(err)
//		return err
//	}
//
//	musdb = s.p.mdb.UTable(qdb.WithContext(context.Background()), a.TableName())
//}

func (s *ServerGD) SearchSongs(query string, page int, mode string) ([]*gdps_db.Song, int, error) {
	a := gdps_db.Song{}
	mus := services.InitMusic(s.p.redis, s.Srv.SrvID+"_search")

	qdb, err := s.p.mdb.OpenMutated("gdps", s.Srv.SrvID)
	defer s.p.mdb.DisposeMutated("gdps", s.Srv.SrvID)
	if utils.Should(err) != nil {
		log.Println(err)
		return nil, 0, err
	}

	qdb = s.p.mdb.UTable(qdb, a.TableName())

	cnt := s.getSongCount(qdb)

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
			if utils.Should(err) != nil {
				continue
			}
			song.URL = xmus.Url
		}
		if strings.Contains(song.URL, "mediapool.halhost.cc") {
			song.URL = strings.ReplaceAll(song.URL, "mediapool.halhost.cc", "cdn2.fruitspace.one")
		}
	}
	return songs, cnt, nil
}

func (s *ServerGD) getSongCount(gdb *gorm.DB) int {
	var cnt int64
	gdb.Count(&cnt)
	return int(cnt)
}

func (s *ServerGD) AddSong(xtype string, url string, meta string) (*gdps_db.Song, error) {
	a := gdps_db.Song{}
	mus := services.InitMusic(s.p.redis, s.Srv.SrvID+"-"+meta)

	qdb, err := s.p.mdb.OpenMutated("gdps", s.Srv.SrvID)
	defer s.p.mdb.DisposeMutated("gdps", s.Srv.SrvID)
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
	id := s.pushSong(qdb, resp, xtype, rid)
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

func (s *ServerGD) pushSong(qdb *gorm.DB, response *structs.MusicResponse, xtype string, rid string) int {
	if f, _ := regexp.MatchString(`[^0-9\.]`, response.Size.String()); f || response.Size == "" {
		response.Size = "5.00"
	}

	sz, _ := response.Size.Float64()
	url := response.Url
	if rid != "" {
		url = fmt.Sprintf("hal:%s:%s", xtype, rid)
	}
	song := gdps_db.Song{
		Name:   response.Name,
		Artist: response.Artist,
		Size:   sz,
		URL:    url,
	}
	var rsong *gdps_db.Song
	qdb.WithContext(context.Background()).Where(gdps_db.Song{URL: song.URL}).First(&rsong)
	if rsong.ID > 0 {
		return rsong.ID
	}
	qdb.WithContext(context.Background()).Create(&song)
	return song.ID
}

//endregion

func (s *ServerGD) UpgradeServer(uid int, srvid string, tariffid int, duration string, promocode string) error {
	pm := NewPromocodeProvider(s.p.db)

	preg := regexp.MustCompile("^[a-zA-Z0-9]+$")
	defer func() {
		// recover from panic if one occured. Set err to nil otherwise.
		v := recover()
		if v != nil {
			log.Println(v, string(debug.Stack()))
		}
	}()
	if !preg.MatchString(srvid) || !s.Exists(srvid) {
		return errors.New("Invalid srvid |srvid")
	}
	//s.GetServerBySrvID(srvid)
	//if uid != s.Srv.OwnerID && uid != 1 {
	//	return errors.New("Invalid owner")
	//}
	// Checked by api AUTH
	s.LoadCoreConfig()

	if tariffid < s.Srv.Plan || tariffid > len(fiberapi.ProductGDTariffs) {
		return errors.New("Invalid Tariff |Tariff")
	}
	tariff := fiberapi.ProductGDTariffs[strconv.Itoa(tariffid)]

	when := time.Now()
	// Select which is latter (only non-free tariffs)
	if when.Compare(s.Srv.ExpireDate) > 0 && s.Srv.Plan > 1 {
		when = s.Srv.ExpireDate
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

	if err := s.p.db.Model(&s.Srv).WhereBinary(db.ServerGd{SrvID: srvid}).Updates(db.ServerGd{ExpireDate: when, Plan: tariffid}).Error; err != nil {
		log.Println(err)
		return errors.New("DATABASE ERROR. REPORT IMMEDIATELY")
	}
	s.CoreConfig.ServerConfig.MaxUsers = tariff.Players
	s.CoreConfig.ServerConfig.MaxLevels = tariff.Levels
	s.CoreConfig.ServerConfig.MaxPosts = tariff.Posts
	s.CoreConfig.ServerConfig.MaxComments = tariff.Comments
	s.CoreConfig.ServerConfig.Locked = false

	vdata, _ := json.Marshal(s.CoreConfig)
	return utils.Should(s.p.redis.Get("gdps").Set(context.Background(), srvid, string(vdata), 0).Err())

}

func (s *ServerGD) CreateServer(uid int, name string, tariffid int, duration string, promocode string) (string, error) {
	pm := NewPromocodeProvider(s.p.db)
	cbs := services.NewBuildService(s.p.db, s.p.mdb, s.p.redis).
		WithConfig(s.p.s3config, s.p.minioconfig).WithAssets(s.p.assets)

	name = strings.TrimSpace(name)
	preg := regexp.MustCompile("^[a-zA-Z0-9 ._-]+$")
	if !preg.MatchString(name) {
		return "", errors.New("Invalid name |name")
	}
	if tariffid < 1 || tariffid > len(fiberapi.ProductGDTariffs) {
		return "", errors.New("Invalid Tariff |Tariff")
	}
	tariff := fiberapi.ProductGDTariffs[strconv.Itoa(tariffid)]
	when := time.Now()
	switch duration {
	case "all":
		when = time.Date(2050, 1, 2, 0, 0, 0, 0, time.UTC)
	case "yr":
		when = when.AddDate(1, 0, 0)
	default:
		when = when.AddDate(0, 1, 0)
	}

	if tariffid == 1 {
		//Temporary
		return "", errors.New("Free server creation is disabled for now |Free server creation is disabled for now")
		var cnt int64
		s.p.db.Model(db.ServerGd{}).Where(db.ServerGd{OwnerID: uid, Plan: 1}).Count(&cnt)
		if cnt != 0 {
			return "", errors.New("You already have FREE server")
		}
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
				return "", errors.New("Invalid promocode |promo_invalid")
			}
			prc, err := promo.Use(price, "gd", strconv.Itoa(tariffid))
			if err != nil {
				return "", err
			}
			price = prc
		}

		price--

		resp := s.p.payments.SpendMoney(uid, price)
		if resp.Status != "ok" {
			return "", errors.New(resp.Message)
		}
	}

	cs := db.ServerGd{
		SrvID:      cbs.CheckAvail("0001"),
		OwnerID:    uid,
		Plan:       tariffid,
		SrvName:    name,
		ExpireDate: when,
	}
	DbPass, SrvKey, err := cbs.InstallServer(cs.SrvID, tariff.Players, tariff.Levels, -1, -1)
	if err != nil {
		return "", err
	}
	cs.DbPassword = DbPass
	cs.SrvKey = SrvKey

	err = s.p.db.Model(db.ServerGd{}).Create(&cs).Error
	if err != nil {
		return "", err
	}
	err = cbs.PushBuildQueue(cs.SrvID, cs.SrvName, "gd_default.png", "2.1", 1, true, false, false,
		"default", "ru", cs.Plan < 2)

	return cs.SrvID, err
}

func (s *ServerGD) UpdateLogo(img []byte) error {
	theImg, _, _ := image.Decode(bytes.NewReader(img))
	cropImg := utils.CropSquareImage(theImg)
	var buf bytes.Buffer
	if err := utils.Should(png.Encode(&buf, cropImg)); err != nil {
		return err
	}
	newImg, _ := io.ReadAll(&buf)

	creds := credentials.NewStaticCredentials(s.p.s3config["access_key"], s.p.s3config["secret"], "")
	cfg := aws.NewConfig().WithEndpoint(s.p.s3config["endpoint"]).WithRegion(s.p.s3config["region"]).WithCredentials(creds)
	sess, err := session.NewSession()
	if utils.Should(err) != nil {
		return err
	}
	svc := s3.New(sess, cfg)
	s.Srv.Icon = "gd_" + s.Srv.SrvID + ".png"

	params := &s3.PutObjectInput{
		Bucket:        aws.String(s.p.s3config["bucket"]),
		Key:           aws.String("server_icons/" + s.Srv.Icon),
		Body:          bytes.NewReader(newImg),
		ContentLength: aws.Int64(int64(len(newImg))),
		ContentType:   aws.String("image/png"),
	}
	_, err = svc.PutObject(params)
	if utils.Should(err) != nil {
		return err
	}
	return s.p.db.Model(s.Srv).Updates(db.ServerGd{Icon: s.Srv.Icon}).Error
}

func (s *ServerGD) UploadTextures(inp io.Reader) error {

	buf := bytes.NewBuffer(nil)
	go func() {
		_, _ = io.Copy(buf, inp)
	}()

	creds := credentials.NewStaticCredentials(s.p.s3config["access_key"], s.p.s3config["secret"], "")
	cfg := aws.NewConfig().WithEndpoint(s.p.s3config["endpoint"]).WithRegion(s.p.s3config["region"]).WithCredentials(creds)
	sess, err := session.NewSession()
	if utils.Should(err) != nil {
		return err
	}
	svc := s3.New(sess, cfg)
	s.Srv.Icon = "gd_" + s.Srv.SrvID + ".png"

	params := &s3.PutObjectInput{
		Bucket:      aws.String(s.p.s3config["bucket"]),
		Key:         aws.String("server_icons/" + s.Srv.Icon),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("application/zip"),
	}
	_, err = svc.PutObject(params)
	if utils.Should(err) != nil {
		return err
	}

	return nil
}

func (s *ServerGD) FetchBuildStatus() string {
	vdb := s.p.db.WithContext(context.Background())
	cbs := services.NewBuildService(vdb, s.p.mdb, s.p.redis).
		WithConfig(s.p.s3config, s.p.minioconfig).WithAssets(s.p.assets)
	return cbs.CheckBuildStatusForGD(s.Srv.SrvID)
}

func (s *ServerGD) ExecuteBuildLab(conf structs.BuildLabSettings) error {
	defer func() {
		// recover from panic if one occured. Set err to nil otherwise.
		v := recover()
		if v != nil {
			log.Println(v, string(debug.Stack()))
		}
	}()
	vdb := s.p.db.WithContext(context.Background())
	if conf.SrvName != "" {
		conf.SrvName = strings.TrimSpace(conf.SrvName)
		preg := regexp.MustCompile("^[a-zA-Z0-9 ._-]+$")
		if !preg.MatchString(conf.SrvName) {
			return errors.New("Invalid name |name")
		}
		s.Srv.SrvName = conf.SrvName
		if err := utils.Should(s.p.db.Model(s.Srv).Updates(db.ServerGd{SrvName: s.Srv.SrvName}).Error); err != nil {
			return err
		}
	} else {
		conf.SrvName = s.Srv.SrvName
	}
	if conf.Version != "2.2" {
		conf.Version = "2.1"
	}
	if conf.Icon != "custom" {
		conf.Icon = "gd_default.png"
	}
	if conf.Textures == "default" {
		s.Srv.IsCustomTextures = false
		s.p.db.Model(s.Srv).UpdateColumn(gorm.Column(db.ServerGd{}, "IsCustomTextures"), 0)
	} else {
		r, err := http.Head(conf.Textures)
		if err != nil || r.StatusCode != 200 {
			merr := ""
			if err != nil {
				merr = err.Error()
			}
			return errors.New("Invalid textures URL:" + conf.Textures + "----" + merr + "==")
		}
	}

	cbs := services.NewBuildService(vdb, s.p.mdb, s.p.redis).
		WithConfig(s.p.s3config, s.p.minioconfig).WithAssets(s.p.assets)

	andro := 0
	if conf.Android {
		andro = 1
	}
	return cbs.PushBuildQueue(s.Srv.SrvID, conf.SrvName, s.Srv.Icon, conf.Version, andro, conf.Windows, conf.IOS, conf.MacOS,
		conf.Textures, "ru", s.Srv.Plan < 2) //! Default textures
}

func (s *ServerGD) DeleteServer() error {
	cbs := services.NewBuildService(s.p.db, s.p.mdb, s.p.redis).
		WithConfig(s.p.s3config, s.p.minioconfig).WithAssets(s.p.assets)

	return cbs.DeleteServer(s.Srv.SrvID, s.Srv.SrvName, s.Srv.Plan < 2)
}

func (s *ServerGD) FreezeServer() {
	if s.LoadCoreConfig() != nil {
		return
	}
	s.CoreConfig.ServerConfig.Locked = true
	vdata, _ := json.Marshal(s.CoreConfig)
	utils.Should(s.p.redis.Get("gdps").Set(context.Background(), s.Srv.SrvID, string(vdata), 0).Err())
	// Who cares
	s.p.db.Model(s.Srv).Updates(db.ServerGd{ExpireDate: time.Now()})
}

func (s *ServerGD) DeleteInstallers() error {
	cbs := services.NewBuildService(s.p.db, s.p.mdb, s.p.redis).
		WithConfig(s.p.s3config, s.p.minioconfig).WithAssets(s.p.assets)

	return cbs.DeleteInstallers(s.Srv.SrvID, s.Srv.SrvName, s.Srv.Plan < 2)
}

func (s *ServerGD) NewGDPSUser() *ServerGDUser {
	return NewServerGDUserSession(s.p, s.Srv.SrvID)
}

func (s *ServerGD) NewInteractor() *ServerGDInteractor {
	return NewServerGDInteractorSession(s.p, s.Srv.SrvID)
}

func (s *ServerGD) SendWebhook(xtype string, data map[string]string) {
	embd := utils.GetEmbed(xtype, data)
	embd.Username = s.Srv.SrvName
	embd.AvatarURL = "https://" + s.p.s3config["cdn"] + "/server_icons/" + s.Srv.Icon
	url, ok := s.Srv.MStatHistory[xtype]
	if !ok {
		log.Println("Not ok")
		log.Printf("%+v", s.Srv.MStatHistory)
		return
	}
	err := embd.SendToWebhook(url.(string))
	if err != nil {
		log.Println(err, url.(string))
	}
}

func (s *ServerGD) ModuleDiscord(enable bool, data map[string]interface{}) error {
	err := s.LoadCoreConfig()
	if err != nil {
		return err
	}
	if s.CoreConfig.ServerConfig.EnableModules == nil {
		s.CoreConfig.ServerConfig.EnableModules = make(map[string]bool)
	}
	s.CoreConfig.ServerConfig.EnableModules["discord"] = enable
	updated, _ := json.Marshal(s.CoreConfig)
	err = s.p.redis.Get("gdps").Set(context.Background(), s.Srv.SrvID, string(updated), 0).Err()
	if utils.Should(err) != nil {
		log.Println(err)
	} else {
		s.Srv.MStatHistory = data
		err = s.p.db.Model(&s.Srv).WhereBinary(db.ServerGd{SrvID: s.Srv.SrvID}).
			Updates(db.ServerGd{MStatHistory: s.Srv.MStatHistory}).Error
	}
	return err
}

//endregion
