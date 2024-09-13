package main

import (
	"database/sql"
	"fmt"
	mysql "github.com/cradio/gorm_mysql"
	"github.com/cradio/gormx"
	fiberapi "github.com/fruitspace/FiberAPI"
	"github.com/fruitspace/FiberAPI/api"
	"github.com/fruitspace/FiberAPI/api/ent"
	_ "github.com/fruitspace/FiberAPI/docs"
	"github.com/fruitspace/FiberAPI/providers"
	"github.com/fruitspace/FiberAPI/providers/ServerGD"
	"github.com/fruitspace/FiberAPI/providers/ServerMC"
	"github.com/fruitspace/FiberAPI/providers/particle"
	"github.com/fruitspace/FiberAPI/utils"
	"github.com/fruitspace/schemas/db/go/db"
	"github.com/getsentry/sentry-go"
	"log"
	"os"
	"time"
)

// @title		FruitSpace FiberAPI
// @version	1.0
// @BasePath	/v2/
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	sentry.Init(sentry.ClientOptions{
		Dsn: "https://ef8c6a708a684aa78fdfc0be5a85115b@o1404863.ingest.sentry.io/4504374313222144",
	})
	defer sentry.Flush(2 * time.Second)

	if len(os.Args) > 1 && os.Args[1] == "-test" {
		api.StartServer(ent.API{Host: ":8080"})
		return
	}

	// Bind Databases
	//newLogger := logger.New(
	//	log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
	//	logger.Config{
	//		LogLevel: logger.Info, // Log level
	//	},
	//)
	DB, err := gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true",
		fiberapi.DB_USER, fiberapi.DB_PASS, fiberapi.DB_HOST, fiberapi.DB_NAME)), &gorm.Config{
		//Logger:                 newLogger,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Println("Error while connecting to " + fiberapi.DB_USER + "@" + fiberapi.DB_HOST + ": " + err.Error())
		//CachedKV.Close()
		time.Sleep(10 * time.Second)
		main()
	}

	if migrateDB(DB) != nil {
		utils.SendMessageDiscord("Database migration failed!")
	}

	//Bind Redis
	Redis := utils.NewMultiRedis().WithDefault(fiberapi.REDIS_HOST, fiberapi.REDIS_PASS).
		Add("sessions", 5).
		Add("gdps", 7).
		Add("music", 8)
	if errs := Redis.Errors(); len(errs) > 0 {
		for _, err := range errs {
			log.Println(err)
		}
		time.Sleep(10 * time.Second)
		main()
	}

	gdpsdbDSN := fiberapi.GDPSDB_USER + ":" + fiberapi.GDPSDB_PASS + "@tcp(" + fiberapi.GDPSDB_HOST + ")/?parseTime=true&multiStatements=true"
	GDPS_DB, _ := sql.Open("mysql", gdpsdbDSN)
	//providers
	accProvider := providers.NewAccountProvider(DB, Redis).
		WithKeys(fiberapi.KEYS, fiberapi.CONFIG, fiberapi.S3_CONFIG).
		WithAssets(&fiberapi.AssetsDir)
	notifProvider := providers.NewNotificationProvider(DB)
	payProvider := providers.NewPaymentProvider(DB, fiberapi.PAYMENTS_HOST)
	promoProvider := providers.NewPromocodeProvider(DB)
	shopProvider := providers.NewShopProvider(DB)
	srvGDProvider := ServerGD.NewServerGDProvider(DB, utils.NewMultiSQL(GDPS_DB, gdpsdbDSN), Redis).
		WithKeys(fiberapi.KEYS, fiberapi.CONFIG, fiberapi.S3_CONFIG, fiberapi.MINIO_CONFIG).
		WithAssets(&fiberapi.AssetsDir).
		WithPaymentsProvider(payProvider)
	srvMCProvider := ServerMC.NewServerMCProvider(DB, payProvider)

	// Consul
	consulKV, err := GetConsulKV()
	if err != nil {
		log.Println("[main] Maintenance and Lock functionality will be degraded. Error: " + err.Error())
		utils.SendMessageDiscord("Maintenance and Lock functionality will be degraded. Error: " + err.Error())
	}

	API := ent.API{
		AccountProvider:      accProvider,
		NotificationProvider: notifProvider,
		PaymentProvider:      payProvider,
		PromocodeProvider:    promoProvider,
		ShopProvider:         shopProvider,
		ServerGDProvider:     srvGDProvider,
		ServerMCProvider:     srvMCProvider,
		ParticleProvider:     particle.NewParticleProvider(DB),
		Host:                 fiberapi.ADDR,

		SuperLock: utils.NewSuperLock(consulKV, SessionID),
	}

	//Consul leadership stuff
	PrepareElection(&API)
	defer StepDown()

	log.Fatalln(api.StartServer(API))
}

func migrateDB(DB *gorm.DB) (err error) {
	if err = DB.AutoMigrate(&db.ACLGd{}); err != nil {
		log.Println(err)
	}
	if err = DB.AutoMigrate(&db.MinecraftNetwork{}); err != nil {
		log.Println(err)
	}
	if err = DB.AutoMigrate(&db.MinecraftServer{}); err != nil {
		log.Println(err)
	}
	if err = DB.AutoMigrate(&db.Pricing{}); err != nil {
		log.Println(err)
	}
	if err = DB.AutoMigrate(&db.Region{}); err != nil {
		log.Println(err)
	}
	if err = DB.AutoMigrate(&db.Tariff{}); err != nil {
		log.Println(err)
	}
	return
}
