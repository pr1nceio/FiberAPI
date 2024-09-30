package ServerGD

import (
	"embed"
	"fmt"
	gorm "github.com/cradio/gormx"
	"github.com/fruitspace/FiberAPI/providers"
	"github.com/fruitspace/FiberAPI/utils"
	"github.com/fruitspace/schemas/db/go/db"
	"time"
)

//region ServerGDProvider

type ServerGDProvider struct {
	db          *gorm.DB
	mdb         *utils.MultiSQL
	redis       *utils.MultiRedis
	payments    *providers.PaymentProvider
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

func (sgp *ServerGDProvider) WithPaymentsProvider(pm *providers.PaymentProvider) *ServerGDProvider {
	sgp.payments = pm
	return sgp
}

func (sgp *ServerGDProvider) WithAssets(assets *embed.FS) *ServerGDProvider {
	sgp.assets = assets
	return sgp
}

func (sgp *ServerGDProvider) New() *ServerGD {
	return &ServerGD{Srv: &db.ServerGD{}, p: sgp}
}

func (sgp *ServerGDProvider) ExposeRedis() *utils.MultiRedis {
	return sgp.redis
}

func (sgp *ServerGDProvider) ExposeGorm() *gorm.DB {
	return sgp.db
}

func (sgp *ServerGDProvider) GetUserServers(uid int) []*db.ServerGDSmall {
	var srvs []*db.ServerGDSmall
	sgp.db.Model(db.ServerGD{}).Where(db.ServerGD{OwnerID: uid}).Find(&srvs)
	for _, srv := range srvs {
		srv.Icon = "https://" + sgp.s3config["cdn"] + "/server_icons/" + srv.Icon
	}

	return srvs
}

func (sgp *ServerGDProvider) GetTopServers(offset int) []*db.ServerGDSmall {
	var srvs []*db.ServerGDSmall
	sgp.db.Model(db.ServerGD{}).Where(fmt.Sprintf("%s>1", gorm.Column(db.ServerGD{}, "Plan"))).
		Where(fmt.Sprintf("%s>NOW()", gorm.Column(db.ServerGD{}, "ExpireDate"))).
		Order(fmt.Sprintf("%s DESC", gorm.Column(db.ServerGD{}, "UserCount"))).
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
	sgp.db.Model(db.ServerGD{}).Count(&cnt)
	return int(cnt)
}

func (sgp *ServerGDProvider) CountLevels() int {
	var cnt int64
	sgp.db.Table((&db.ServerGD{}).TableName()).Select(fmt.Sprintf("sum(%s)", gorm.Column(db.ServerGD{}, "LevelCount"))).Row().Scan(&cnt)
	return int(cnt)
}

func (sgp *ServerGDProvider) GetUnpaidServers() []string {
	var srvs []*db.ServerGDSmall
	var srvids []string
	sgp.db.Model(db.ServerGD{}).Where(fmt.Sprintf("%s<NOW()", gorm.Column(db.ServerGD{}, "ExpireDate"))).Find(&srvs)
	for _, srv := range srvs {
		srvids = append(srvids, srv.SrvID)
	}
	return srvids
}

func (sgp *ServerGDProvider) GetInactiveServers(maxUsers int, free bool) []string {
	var srvs []*db.ServerGDSmall
	var srvids []string
	tx := sgp.db.Model(db.ServerGD{}).
		Where(fmt.Sprintf("%s<(CURRENT_DATE - INTERVAL 7 DAY)", gorm.Column(db.ServerGD{}, "CreatedAt"))).
		Where(fmt.Sprintf("%s<=%d", gorm.Column(db.ServerGD{}, "UserCount"), maxUsers))
	if free {
		tx = tx.Where(db.ServerGD{Plan: 1})
	}
	tx.Find(&srvs)
	for _, srv := range srvs {
		srvids = append(srvids, srv.SrvID)
	}
	return srvids
}

func (sgp *ServerGDProvider) GetMissingInstallersServers() []string {
	var srvs []*db.ServerGDSmall
	var srvids []string
	tx := sgp.db.Model(db.ServerGD{}).
		//Where(fmt.Sprintf("%s<(CURRENT_DATE - INTERVAL 1 DAY)", gorm.Column(db.ServerGD{}, "CreatedAt"))).
		Where(fmt.Sprintf("%s=''", gorm.Column(db.ServerGD{}, "ClientWindowsURL")))
	tx.Find(&srvs)
	for _, srv := range srvs {
		srvids = append(srvids, srv.SrvID)
	}
	return srvids
}

//endregion
