package main

import (
	"fmt"
	"github.com/fruitspace/FiberAPI/api/ent"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/fruitspace/FiberAPI/services"
	"github.com/fruitspace/FiberAPI/utils"
	"github.com/go-co-op/gocron"
	"log"
	"strconv"
)
import consul "github.com/hashicorp/consul/api"

var ucron = gocron.NewScheduler(utils.Loc)
var conn *ent.API

var LEADER = false
var SessionID string
var KvEngine *consul.KV

func MaintainTasksDaily() {
	if !LEADER {
		return
	}

	utils.SendMessageDiscord("Starting maintenance...")

	// Check paid GDPS servers expiry

	{
		gdpslist := conn.ServerGDProvider.GetUnpaidServers()
		freezeReport := "### Freezing Paid GDPS (" + strconv.Itoa(len(gdpslist)) + ")...\n"
		for _, gdps := range gdpslist {
			srv := conn.ServerGDProvider.New()
			if !srv.GetServerBySrvID(gdps) {
				continue
			}
			srv.LoadCoreConfig()
			if srv.CoreConfig.ServerConfig.Locked {
				continue
			}
			srv.FreezeServer()
			freezeReport += fmt.Sprintf("\n❄️ %s is frozen", gdps)
			if len(freezeReport) > 500 {
				utils.SendMessageDiscord(freezeReport)
				freezeReport = ""
			}
		}
		utils.SendMessageDiscord(freezeReport)
	}

	// Purge Empty Free GDPS and Freeze them
	{
		gdpslist := conn.ServerGDProvider.GetInactiveServers(3, true)
		freezeReport := "### Purging Free Inactive GDPS (" + strconv.Itoa(len(gdpslist)) + ")...\n"
		for _, gdps := range gdpslist {
			srv := conn.ServerGDProvider.New()
			if !srv.GetServerBySrvID(gdps) {
				continue
			}
			srv.LoadCoreConfig()
			interactor := srv.NewInteractor()
			usrs := interactor.CountActiveUsersLastWeek()

			if usrs > 0 {
				continue
			}
			// If there's no active users last week then freeze server and set expire date to current time to ensure removal in future

			if srv.CoreConfig.ServerConfig.Locked {
				continue
			}
			srv.FreezeServer()
			err := srv.DeleteInstallers()
			if err != nil {
				freezeReport += fmt.Sprintf("\n❌ Couldn't purge %s, error: `%s`", gdps, err.Error())
			} else {
				freezeReport += fmt.Sprintf("\n❄️ %s is purged", gdps)
			}
			if len(freezeReport) > 500 {
				utils.SendMessageDiscord(freezeReport)
				freezeReport = ""
			}
		}
	}

	// Fix missing installers
	{
		gdpslist := conn.ServerGDProvider.GetMissingInstallersServers()
		freezeReport := "### Restoring GDPS with missing installers (" + strconv.Itoa(len(gdpslist)) + ")...\n"
		for _, gdps := range gdpslist {
			srv := conn.ServerGDProvider.New()
			if !srv.GetServerBySrvID(gdps) {
				continue
			}
			srv.LoadCoreConfig()

			if srv.CoreConfig.ServerConfig.Locked {
				continue
			}
			err := srv.ExecuteBuildLab(structs.BuildLabSettings{
				SrvName:  srv.Srv.SrvName,
				Version:  "2.1",
				Windows:  true,
				Android:  true,
				IOS:      false,
				MacOS:    false,
				Icon:     "gd_default.png",
				Textures: "default",
			})
			if err != nil {
				freezeReport += fmt.Sprintf("\n❌ Couldn't recover %s, error: `%s`", gdps, err.Error())
			} else {
				freezeReport += fmt.Sprintf("\n❄⚙️ %s is recovered W+A", gdps)
			}
			if len(freezeReport) > 500 {
				utils.SendMessageDiscord(freezeReport)
				freezeReport = ""
			}
		}
		utils.SendMessageDiscord(freezeReport)
	}

	//Clear music
	mus := services.InitMusic(conn.ServerGDProvider.ExposeRedis(), "admin")
	mcnt := mus.CleanEmptyNewgrounds()
	utils.SendMessageDiscord(fmt.Sprintf("Cleaned %d invalid NG songs. \n### Maintenance Complete", mcnt))
}

func GetConsulKV() (consulKV *consul.KV, err error) {
	consulConf := consul.DefaultConfig()
	consulConf.Address = utils.GetEnv("CONSUL_ADDR", "127.0.0.1")
	consulConf.Token = utils.GetEnv("CONSUL_TOKEN", "")
	consulConf.Datacenter = utils.GetEnv("CONSUL_DC", "m41")
	consulCli, err := consul.NewClient(consulConf)
	if err != nil {
		log.Println("Unable to connect to Consul cluster. Assuming self-leadership: " + err.Error())
		return nil, err
	}
	KvEngine = consulCli.KV()
	SessID, _, err := consulCli.Session().Create(&consul.SessionEntry{Name: "FiberAPI", TTL: "5m"}, nil)
	SessionID = SessID
	if err != nil {
		log.Println("Unable to connect to create Consul Session. Assuming self-leadership: " + err.Error())
		return nil, err
	}
	ucron.Every(30).Seconds().Do(func() {
		consulCli.Session().Renew(SessID, nil)
	})
	return KvEngine, nil
}

func PrepareElection(iconn *ent.API) {
	KvEngine = iconn.SuperLock.ExposeConsul()
	conn = iconn
	if KvEngine == nil {
		log.Println("Unable to connect to create Session. Assuming self-leadership")
		LEADER = true
	} else {
		AquireLeadership()
		if !LEADER {
			log.Println("Couldn't acquire leadership. Dispatching 10sec watchdog")
			if _, err := ucron.Every(10).Seconds().Do(AquireLeadership); err != nil {
				log.Println(err)
			}
		}
	}
	_, err := ucron.Every(1).Day().At("00:00").Do(MaintainTasksDaily)
	if err != nil {
		log.Println("CANNOT LAUNCH TASKS")
	}
	ucron.StartAsync()

}

func AquireLeadership() {
	kvData := &consul.KVPair{
		Key:     "sessions/fiberapi_lead",
		Value:   []byte(utils.GetEnv("NOMAD_SHORT_ALLOC_ID", "default")),
		Session: SessionID,
	}
	isAcq, _, err := KvEngine.Acquire(kvData, nil)
	if err == nil && isAcq {
		if LEADER {
			log.Println("Still leader (ensuring tasks)")
		} else {
			log.Println("Lock was successfully acquired. NOW LEADER")
			LEADER = true
		}
	} else {
		log.Println(err)
		if LEADER {
			log.Println("Couldn't acquire leadership. Stepped down by force.")
			LEADER = false
		} else {
			log.Println("Couldn't acquire leadership. Still follower")
		}
	}

}

func StepDown() {
	kvData := &consul.KVPair{
		Key:     "sessions/ghostcore_lead",
		Value:   []byte(utils.GetEnv("NOMAD_SHORT_ALLOC_ID", "default")),
		Session: SessionID,
	}
	isRel, _, err := KvEngine.Release(kvData, nil)
	if err == nil && isRel {
		log.Println("Lock was successfully released. NOW FOLLOWER")
		LEADER = false
	} else {
		log.Println("[!!!] COULD NOT RELEASE LOCK [!!!]")
	}
}
