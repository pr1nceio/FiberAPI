package cmd

import (
	"database/sql"
	"fmt"
	"github.com/cradio/gorm_mysql"
	"github.com/cradio/gormx"
	fiberapi "github.com/fruitspace/FiberAPI"
	"github.com/fruitspace/FiberAPI/providers"
	"github.com/fruitspace/FiberAPI/utils"
	"github.com/getsentry/sentry-go"
	"log"
	"time"
)

var (
	DB    *gorm.DB
	Redis *utils.MultiRedis
)

func main() {
	sentry.Init(sentry.ClientOptions{
		Dsn: "https://ef8c6a708a684aa78fdfc0be5a85115b@o1404863.ingest.sentry.io/4504374313222144",
	})
	defer sentry.Flush(2 * time.Second)

	// Bind Databases
	var err error
	DB, err = gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s)/%s",
		fiberapi.DB_USER, fiberapi.DB_PASS, fiberapi.DB_HOST, fiberapi.DB_NAME)))
	if err != nil {
		log.Println("Error while connecting to " + fiberapi.DB_USER + "@" + fiberapi.DB_HOST + ": " + err.Error())
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

	xrdb, _ := sql.Open("mysql", fiberapi.DB_USER+":"+fiberapi.DB_PASS+"@tcp("+fiberapi.DB_HOST+")/default_db")
	xs := providers.NewServerGDProvider(DB, utils.NewMultiSQL(xrdb), nil).New()

	r1, r2, r3 := xs.GetLogs(-1, 0)
	log.Printf("%+v, %d, %+v", r1, r2, r3)

}
