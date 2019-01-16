package main

import (
	"encoding/json"
	"fmt"
	"github.com/camsiabor/qcom/agenda"
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/qdao"
	"github.com/camsiabor/qcom/qlog"
	"github.com/camsiabor/qcom/qref"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qdaobundle/qelastic"
	"github.com/camsiabor/qdaobundle/qredis"
	"github.com/camsiabor/qstock/dict"
	"github.com/camsiabor/qstock/httpv"
	"github.com/camsiabor/qstock/run/rscript"
	"github.com/camsiabor/qstock/sync"
	"github.com/camsiabor/qstock/sync/maincache"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"time"
)

func master(g *global.G) {

	go initPerfAnalysis(g)

	var jsonstr, _ = json.Marshal(g.Config)
	qlog.Log(qlog.INFO, "config", string(jsonstr[:]))

	var master_listen = util.GetStr(g.Config, "127.0.0.1:65000", "master", "listen")
	if !strings.Contains(master_listen, ":") {
		master_listen = ":" + master_listen
	}
	var lerr error
	g.Listener, lerr = net.Listen("tcp", master_listen)
	if lerr != nil {
		qlog.Log(qlog.INFO, "establish master listen fail ", master_listen, lerr)
		panic(lerr)
	}
	qlog.Log(qlog.INFO, "master", "listener establish", master_listen)

	var timezone = util.GetStr(g.Config, "Asia/Shanghai", "global", "timezone")
	time.LoadLocation(timezone)

	// [mapper] ------------------------------------------------------------------------------------------------
	var mapperConfig = util.GetMap(g.Config, true, "mapping")
	qref.GetMapperManager().Init(mapperConfig)
	g.SetData("mapperm", qref.GetMapperManager())

	// [dao] ------------------------------------------------------------------------------------------------
	initDao(g)

	// [script] ------------------------------------------------------------------------------------------------
	rscript.InitScript(g)

	// [gin] ------------------------------------------------------------------------------------------------
	httpv.GetInstance().Run()

	// [agenda] ------------------------------------------------------------------------------------------------
	var agendaConfig = util.GetMap(g.Config, true, "agenda")
	agenda.GetAgendaManager().Init(agendaConfig)

	maincache.InitMainCache(g)

	initSyncer(g)

}

func initPerfAnalysis(g *global.G) {
	var config = util.GetMap(g.Config, false, "debug", "http")
	if config == nil {
		return
	}
	var active = util.GetBool(config, false, "active")
	if !active {
		return
	}
	var endpoint = util.GetStr(config, ":8080", "endpoint")
	http.ListenAndServe(endpoint, nil)
}

func initDao(g *global.G) {
	var schemaOpts = util.GetMap(g.Config, true, "dbschema")
	var databaseOpts = util.GetMap(g.Config, true, "database")
	qdao.GetManager().Init(func(manager *qdao.DaoManager, name string, opts map[string]interface{}) (qdao.D, error) {
		var daotype = util.GetStr(opts, "", "type")
		if len(daotype) == 0 {
			panic(fmt.Errorf("no dao type specific in dao init : %v", opts))
		}
		switch daotype {
		case "redis":
			var redis = &qredis.DaoRedis{}
			redis.Framework = manager.GetSchema()
			return redis, nil
		case "elasticsearch":
			var elastic = &qelastic.DaoElastic{}
			return elastic, nil
		default:
			panic("db not implement : " + daotype)
		}
		return nil, nil
	}, schemaOpts, databaseOpts)
	g.SetData("daom", qdao.GetManager())
}

func initSyncer(g *global.G) {
	var api_config = util.GetMap(g.Config, true, "api")
	for apiname, api := range api_config {
		var active = util.GetBool(api, true, "active")
		if !active {
			continue
		}
		var syncer = new(sync.Syncer)
		g.CmdHandlerRegister(dict.SERVICE_SYNC, syncer)
		syncer.Run(apiname)
	}
}
