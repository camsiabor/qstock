package main

// https://www.showapi.com/api/view/131
// http://godoc.org/github.com/go-redis/redis

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"qcom/agenda"
	"qcom/global"
	"qcom/qdao"
	"qcom/scache"
	"qcom/util"
	"qcom/util/qlog"
	"stock/dict"
	"stock/httpv"
	"stock/sync"
	"time"
)

func handleCmd() {
	var chCmd = make(chan string, 256);
	var g = global.GetInstance();
	for {
		var cmd, ok = <- chCmd;
		if (!ok || cmd == "exit") {
			qlog.Log(qlog.INFO, "main", "exit");
			break;
		}
		qlog.Log(qlog.INFO, "main", "receive cmd", cmd);
		if (cmd == "config reload") {
			var config, err = util.ConfigLoad(g.ConfigPath, "includes");
			if (err != nil) {
				qlog.Log(qlog.FATAL, "config", "load failure", err);
			} else {
				g.Config = config;
			}
		}
	}
}


func initSyncer(g * global.Global) {

	var scache_snapshot = scache.GetCacheManager().Get(dict.CACHE_STOCK_SNAPSHOT);
	var scache_khistory = scache.GetCacheManager().Get(dict.CACHE_STOCK_KHISTORY);
	scache_snapshot.Loader = func(scache *scache.SCache, keys []string) (interface{}, error) {
		conn, err := qdao.GetDaoManager().Get(dict.DAO_MAIN);
		if (err != nil) {
			return nil, err;
		}
		var code = keys[0];
		return conn.Get(dict.DB_DEFAULT, "", code, true);
	}
	scache_khistory.Loader = func(scache *scache.SCache, keys []string) (interface{}, error) {
		conn, err := qdao.GetDaoManager().Get(dict.DAO_MAIN);
		if (err != nil) {
			return nil, err;
		}
		var code = keys[0];
		var datestr = keys[1];
		return conn.Get(dict.DB_HISTORY, code, datestr, true);
	};

	var api_config = util.GetMap(g.Config, true, "api");
	for apiname, api := range api_config {
		var active = util.GetBool(api, true, "active");
		if (!active) {
			continue;
		}
		var syncer = new(sync.Syncer);
		g.CmdHandlerRegister(apiname, syncer);
		syncer.Run(apiname);
		//var requester = util.GetStr(api, "", "requester");
		//if (len(requester) > 0) {
		//	var vsyncer = reflect.ValueOf(syncer);
		//	var vrequester = vsyncer.MethodByName(requester);
		//	var frequester = vrequester.Interface();
		//	syncer.RequestHandler = frequester.(sync.SyncRequestHandler);
		//}
	}
}

func main() {

	var g = global.GetInstance();
	g.Continue = true;
	g.PanicHandler = func(pan interface{}) {
		qlog.Log(qlog.ERROR, pan);
	}
	g.Run();

	flag.StringVar(&g.LogPath,  "log", "log", "log file path");
	flag.StringVar(&g.ConfigPath,  "config", "config.json", "configuration file path");
	flag.StringVar(&g.TimeZone,  "timezone", "Asia/Shanghai", "timezone");
	var doHelp = flag.Bool( "help", false, "help");
	var doVersion = flag.Bool( "version", false, "version");
	flag.Parse();
	if (*doVersion) {
		fmt.Println("version");
	}
	if (*doHelp) {
		flag.PrintDefaults();
		os.Exit(0);
	}
	qlog.LogInit(g.LogPath, qlog.INFO, log.Ltime, true);

	qlog.Log(qlog.INFO, "main", "init")
	time.LoadLocation(g.TimeZone)

	// [Config] ------------------------------------------------------------------------------------------------
	if (len(g.ConfigPath) == 0) {
		g.ConfigPath = "config.json"
	}
	var config, err = util.ConfigLoad(g.ConfigPath, "includes")
	if err != nil {
		qlog.Log(qlog.FATAL, "config", "load failure", g.ConfigPath, err)
		return
	}
	g.Config = config;
	var jsonstr, _ = json.Marshal(config)
	qlog.Log(qlog.INFO, "config", string(jsonstr[:]));

	var timezone = util.GetStr(g.Config, "Asia/Shanghai", "global", "timezone");
	time.LoadLocation(timezone)

	// [mapper] ------------------------------------------------------------------------------------------------
	var mapperConfig = util.GetMap(g.Config, true, "mapping");
	util.GetMapperManager().Init(mapperConfig);

	// [redis] ------------------------------------------------------------------------------------------------
	qdao.GetDaoManager().Init();

	// [gin] ------------------------------------------------------------------------------------------------
	httpv.GetInstance().Run();

	// [agenda] ------------------------------------------------------------------------------------------------
	var agendaConfig = util.GetMap(g.Config, true, "agenda");
	agenda.GetAgendaManager().Init(agendaConfig);

	// [api] --------------------------------------------------------------------------------------------
	initSyncer(g);


	// [cmd] --------------------------------------------------------------------------------------------
	handleCmd();

	// [release] --------------------------------------------------------------------------------------------
	//Redis_destroy();

	qlog.Log(qlog.INFO, "main", "fin")

}
