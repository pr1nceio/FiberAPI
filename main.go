package fiberapi

import (
	"database/sql"
	"embed"
	"fmt"
	"github.com/cradio/gorm_mysql"
	"github.com/cradio/gormx"
	"github.com/fruitspace/FiberAPI/providers"
	"github.com/fruitspace/FiberAPI/utils"
	"github.com/getsentry/sentry-go"
	"log"
	"time"
)

var (
	DB    *gorm.DB
	Redis *utils.MultiRedis

	KEYS      = utils.GetKVEnv("KEYS")      //key_enc=,key_void=
	CONFIG    = utils.GetKVEnv("CONFIG")    //ipinfo_key=,email=,email_pass=,email_host=,hCaptchaToken=
	S3_CONFIG = utils.GetKVEnv("S3_CONFIG") //access_key=,secret=,region=,bucket=,endpoint=,cdn=

	DB_HOST = utils.GetEnv("DB_HOST", "localhost:3306")
	DB_NAME = utils.GetEnv("DB_NAME", "default_db")
	DB_USER = utils.GetEnv("DB_USER", "root")
	DB_PASS = utils.GetEnv("DB_PASS", "root")

	REDIS_HOST = utils.GetEnv("REDIS_HOST", "localhost:6379")
	REDIS_PASS = utils.GetEnv("REDIS_PASS", "")

	PAYMENTS_HOST = utils.GetEnv("PAYMENTS_HOST", "http://localhost:8000")

	ADDR = utils.GetEnv("ADDR", "localhost:8080")
)

//go:embed assets
var assets embed.FS

func main() {
	sentry.Init(sentry.ClientOptions{
		Dsn: "https://ef8c6a708a684aa78fdfc0be5a85115b@o1404863.ingest.sentry.io/4504374313222144",
	})
	defer sentry.Flush(2 * time.Second)

	// Bind Databases
	var err error
	DB, err = gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s)/%s", DB_USER, DB_PASS, DB_HOST, DB_NAME)))
	if err != nil {
		log.Println("Error while connecting to " + DB_USER + "@" + DB_HOST + ": " + err.Error())
		//CachedKV.Close()
		time.Sleep(10 * time.Second)
		main()
	}

	//Bind Redis
	//Redis = utils.NewMultiRedis().WithDefault(REDIS_HOST, REDIS_PASS).
	//	Add("sessions", 5).
	//	Add("gdps", 7).
	//	Add("music", 8)
	//if errs := Redis.Errors(); len(errs) > 0 {
	//	for _, err := range errs {
	//		log.Println(err)
	//	}
	//	time.Sleep(10 * time.Second)
	//	main()
	//}

	// Consul leadership stuff
	//PrepareElection()
	//defer StepDown()

	xrdb, _ := sql.Open("mysql", DB_USER+":"+DB_PASS+"@tcp("+DB_HOST+")/default_db")
	xs := providers.NewServerGDProvider(DB, utils.NewMultiSQL(xrdb), nil).New()

	r1, r2, r3 := xs.GetLogs(-1, 0)
	log.Printf("%+v, %d, %+v", r1, r2, r3)

}
