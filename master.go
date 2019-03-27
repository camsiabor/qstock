package main

import (
	"encoding/json"
	"fmt"
	"github.com/camsiabor/qcom/agenda"
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/qdao"
	"github.com/camsiabor/qcom/qlog"
	"github.com/camsiabor/qcom/qnet"
	"github.com/camsiabor/qcom/qref"
	"github.com/camsiabor/qcom/util"
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

	var master_listen_endpoing = util.GetStr(g.Config, "127.0.0.1:65000", "master", "listen")
	if !strings.Contains(master_listen_endpoing, ":") {
		master_listen_endpoing = ":" + master_listen_endpoing
	}
	var lerr error
	g.Listener, lerr = net.Listen("tcp", master_listen_endpoing)
	if lerr != nil {
		qlog.Log(qlog.INFO, "establish master listen fail ", master_listen_endpoing, lerr)
		panic(lerr)
	}
	qlog.Log(qlog.INFO, "master", "listener establish", master_listen_endpoing)

	var timezone = util.GetStr(g.Config, "Asia/Shanghai", "global", "timezone")
	time.LoadLocation(timezone)

	initMapper(g)

	initDao(g)

	initScript(g)

	initHttp(g)

	initAgenda(g)

	initCache(g)

	initSyncer(g)

	initHttpClient(g)
	/*
		go func() {
			for {
				qlog.Log(qlog.INFO, "hello world")
				time.Sleep(time.Second * 5)
			}
		} ()
	*/
}
func initAgenda(g *global.G) {
	var agendaConfig = util.GetMap(g.Config, true, "agenda")
	agenda.GetAgendaManager().Init(agendaConfig)
}
func initCache(g *global.G) {
	maincache.InitMainCache(g)
}
func initHttp(g *global.G) {
	var httpserv = httpv.GetInstance()
	httpserv.Run()
	g.RegisterModule("httpv", httpserv)
	g.SetData("http", qnet.GetSimpleHttp())
}
func initMapper(g *global.G) {
	var mapperConfig = util.GetMap(g.Config, true, "mapping")
	qref.GetMapperManager().Init(mapperConfig)
	g.SetData("mapperm", qref.GetMapperManager())
}
func initScript(g *global.G) {
	rscript.InitScript(g)
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
	var daoManager = qdao.GetManager()
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
		//case "elasticsearch":
		//	var elastic = &qelastic.DaoElastic{}
		//	return elastic, nil
		default:
			panic("db not implement : " + daotype)
		}
		return nil, nil
	}, schemaOpts, databaseOpts)
	g.RegisterModule("dao_manager", daoManager)
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
		g.RegisterModule("syncer."+apiname, syncer)
	}
}

func initHttpClient(g *global.G) {

	var simpleHttp = qnet.GetSimpleHttp()
	g.SetData("http", simpleHttp)

	var seleniumConfig = util.GetMap(g.Config, true, "httpclient")

	for driverName, driverConfig := range seleniumConfig {
		var httpagent = &httpv.HttpAgent{}
		httpagent.Name = driverName
		httpagent.Config = util.AsMap(driverConfig, true)
		httpagent.InitParameters(httpagent.Config)
		g.RegisterModule(driverName, httpagent)
		g.SetData(driverName, httpagent)
	}

}
