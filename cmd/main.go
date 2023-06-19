package main

import (
	"database/sql"
	"fmt"
	mysql "github.com/cradio/gorm_mysql"
	"github.com/cradio/gormx"
	"github.com/cradio/gormx/logger"
	fiberapi "github.com/fruitspace/FiberAPI"
	"github.com/fruitspace/FiberAPI/api"
	_ "github.com/fruitspace/FiberAPI/docs"
	"github.com/fruitspace/FiberAPI/providers"
	"github.com/fruitspace/FiberAPI/utils"
	"github.com/getsentry/sentry-go"
	"log"
	"os"
	"time"
)

// @title		FruitSpace FiberAPI
// @version	1.0
// @BasePath	/
func main() {
	sentry.Init(sentry.ClientOptions{
		Dsn: "https://ef8c6a708a684aa78fdfc0be5a85115b@o1404863.ingest.sentry.io/4504374313222144",
	})
	defer sentry.Flush(2 * time.Second)

	if len(os.Args) > 1 && os.Args[1] == "-test" {
		api.StartServer(api.API{Host: ":8080"})
		return
	}

	// Bind Databases
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			LogLevel: logger.Info, // Log level
		},
	)
	DB, err := gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true",
		fiberapi.DB_USER, fiberapi.DB_PASS, fiberapi.DB_HOST, fiberapi.DB_NAME)), &gorm.Config{Logger: newLogger})
	if err != nil {
		log.Println("Error while connecting to " + fiberapi.DB_USER + "@" + fiberapi.DB_HOST + ": " + err.Error())
		//CachedKV.Close()
		time.Sleep(10 * time.Second)
		main()
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

	GDPS_DB, _ := sql.Open("mysql", fiberapi.GDPSDB_USER+":"+fiberapi.GDPSDB_PASS+"@tcp("+fiberapi.GDPSDB_HOST+")/default_db?parseTime=true")

	//providers
	accProvider := providers.NewAccountProvider(DB, Redis).
		WithKeys(fiberapi.KEYS, fiberapi.CONFIG, fiberapi.S3_CONFIG).
		WithAssets(&fiberapi.AssetsDir)
	notifProvider := providers.NewNotificationProvider(DB)
	payProvider := providers.NewPaymentProvider(DB, fiberapi.PAYMENTS_HOST)
	promoProvider := providers.NewPromocodeProvider(DB)
	shopProvider := providers.NewShopProvider(DB)
	srvGDProvider := providers.NewServerGDProvider(DB, utils.NewMultiSQL(GDPS_DB), Redis).
		WithKeys(fiberapi.KEYS, fiberapi.CONFIG, fiberapi.S3_CONFIG, fiberapi.MINIO_CONFIG).
		WithAssets(&fiberapi.AssetsDir).
		WithPaymentsProvider(payProvider)

	API := api.API{
		AccountProvider:      accProvider,
		NotificationProvider: notifProvider,
		PaymentProvider:      payProvider,
		PromocodeProvider:    promoProvider,
		ShopProvider:         shopProvider,
		ServerGDProvider:     srvGDProvider,
		Host:                 fiberapi.ADDR,
	}

	//Consul leadership stuff
	PrepareElection(&API)
	defer StepDown()

	log.Fatalln(api.StartServer(API))
}
