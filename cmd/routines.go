package main

import (
	"fmt"
	"github.com/fruitspace/FiberAPI/api"
	"github.com/fruitspace/FiberAPI/services"
	"github.com/fruitspace/FiberAPI/utils"
	"github.com/go-co-op/gocron"
	"log"
)
import consul "github.com/hashicorp/consul/api"

var ucron = gocron.NewScheduler(utils.Loc)
var conn *api.API

var LEADER = false
var SessionID string
var KvEngine *consul.KV

func MaintainTasksDaily() {
	if !LEADER {
		return
	}
	// Check paid GDPS servers expiry
	utils.SendMessageDiscord("Starting maintenance...")
	gdpslist := conn.ServerGDProvider.GetUnpaidServers()
	freezeReport := ""
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

	//Clear music
	mus := services.InitMusic(conn.ServerGDProvider.ExposeRedis())
	mus.CleanEmptyNewgrounds()
}

func PrepareElection(iconn *api.API) {
	conn = iconn
	consulConf := consul.DefaultConfig()
	consulConf.Address = utils.GetEnv("CONSUL_ADDR", "127.0.0.1")
	consulConf.Token = utils.GetEnv("CONSUL_TOKEN", "")
	consulConf.Datacenter = "hal"
	consulCli, err := consul.NewClient(consulConf)
	if err != nil {
		log.Println("Unable to connect to Consul cluster. Assuming self-leadership: " + err.Error())
		LEADER = true
	} else {
		KvEngine = consulCli.KV()
		SessID, _, err := consulCli.Session().Create(&consul.SessionEntry{Name: "FiberAPI", TTL: "5m"}, nil)
		ucron.Every(30).Seconds().Do(func() {
			consulCli.Session().Renew(SessID, nil)
		})
		if err != nil {
			log.Println("Unable to connect to create Session. Assuming self-leadership: " + err.Error())
			LEADER = true
		} else {
			SessionID = SessID
			AquireLeadership()
			if !LEADER {
				log.Println("Couldn't acquire leadership. Dispatching 10sec watchdog")
				if _, err = ucron.Every(10).Seconds().Do(AquireLeadership); err != nil {
					log.Println(err)
				}
			}
		}
	}
	_, err = ucron.Every(1).Day().At("00:00").Do(MaintainTasksDaily)
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
			_, _ = ucron.Every(1).Day().At("00:00").Do(MaintainTasksDaily)
		}
	} else {
		if LEADER {
			log.Println("Couldn't acquire leadership. Stepped down by force.")
			LEADER = false
			ucron.Remove(MaintainTasksDaily)
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
