package main

import (
	"os"
	"os/signal"
	"runtime/pprof"
	"time"

	"github.com/yangsongfwd/backup/log"
	"github.com/yangsongfwd/ddns/config"
	"github.com/yangsongfwd/ddns/data"
	"github.com/yangsongfwd/ddns/data/sqlite"
	"github.com/yangsongfwd/ddns/ddns"
	"github.com/yangsongfwd/ddns/ddns/slave"
	"github.com/yangsongfwd/ddns/download"
	"github.com/yangsongfwd/ddns/server"
	"github.com/yangsongfwd/ddns/web"
)

func regBeforStart() {
	//reg db
	server.Server.RegBeforStart(func() {
		var db data.IDatabase
		db = sqlite.NewSqlite()
		if err := db.Init(); err != nil {
			log.Error(err.Error())
		} else {
			server.Server.AddGlobalData("db", db)
			server.Server.RegBeforStop(func() { db.Close() })
		}
	})

	//reg dns server
	if config.Data.Server.EnableDNS {
		server.Server.RegBeforStart(func() {
			idb := server.Server.GetGlobalData("db")
			if idb == nil {
				log.Error("get db error")
				return
			}

			db := idb.(data.IDatabase)
			var ds *ddns.Server
			ds = &ddns.Server{
				Host:     config.Data.Server.Addr,
				Port:     config.Data.Server.Port,
				RTimeout: 5 * time.Second,
				WTimeout: 5 * time.Second,
				Db:       db,
			}
			ds.Run()

			if config.Data.Server.Master == false {
				s := slave.SlaveServer{}
				err := s.Init(db)
				if err != nil {
					log.Error("init slave failed: %v", err)
				} else {
					s.Start()
				}
			}

			server.Server.AddGlobalData("dnsServer", ds)
			server.Server.RegBeforStop(ds.Stop)
		})
	}

	//reg download mgr
	if config.Data.Download.Enable {
		server.Server.RegBeforStart(func() {
			dload := download.NewDownloadMgr()
			dload.Start()

			server.Server.AddGlobalData("downloadMgr", dload)
		})
	}

	//reg web server
	if config.Data.Server.EnableWeb {
		server.Server.RegBeforStart(func() {
			if config.Data.Web.Admin == "" || config.Data.Web.Passwd == "" {
				log.Error("web admin & passwd can't be empty")
				return
			}

			idb := server.Server.GetGlobalData("db")
			if idb == nil {
				log.Error("get db error")
				return
			}
			db := idb.(data.IDatabase)

			idload := server.Server.GetGlobalData("downloadMgr")
			if idload == nil {
				log.Error("get downloadMgr error")
				return
			}
			dload := idload.(*download.DownloadMgr)

			ws := &web.WebServer{}
			ws.Start(db, dload)
			server.Server.AddGlobalData("ws", ws)
			server.Server.RegBeforStop(ws.Stop)
		})
	}
}

func main() {

	regBeforStart()

	server.Server.Run()

	if config.Data.Server.Debug {
		go profileCPU()
		go profileMEM()
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

forever:
	for {
		select {
		case <-sig:
			log.Debug("signal received, stopping")

			server.Server.Stop()

			break forever
		}
	}
}

func profileCPU() {
	f, err := os.Create("ddns.cprof")
	if err != nil {
		log.Error("%v", err)
		return
	}

	pprof.StartCPUProfile(f)
	time.AfterFunc(6*time.Minute, func() {
		pprof.StopCPUProfile()
		f.Close()
	})
}

func profileMEM() {
	f, err := os.Create("ddns.mprof")
	if err != nil {
		log.Error("%v", err)
		return
	}

	time.AfterFunc(5*time.Minute, func() {
		pprof.WriteHeapProfile(f)
		f.Close()
	})
}
