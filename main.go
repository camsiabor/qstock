package main


/*
// TODO daemon process
// TODO actor, proactor
// TODO micro service
// TODO distribute
// TODO elasticsearch mongodb
 */

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/camsiabor/qcom/agenda"
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/qdao"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util/qconfig"
	"github.com/camsiabor/qcom/util/qerr"
	"github.com/camsiabor/qcom/util/qlog"
	"github.com/camsiabor/qcom/util/qref"
	"github.com/camsiabor/qcom/util/util"
	"github.com/camsiabor/qstock/dict"
	"github.com/camsiabor/qstock/httpv"
	"github.com/camsiabor/qstock/sync"
	"log"
	"os"
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
			var config, err = qconfig.ConfigLoad(g.ConfigPath, "includes");
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
	scache_snapshot.Dao = dict.DAO_MAIN;
	scache_snapshot.Db = dict.DB_DEFAULT;
	scache_khistory.Dao = dict.DAO_MAIN;
	scache_khistory.Db = dict.DB_HISTORY;
	scache_snapshot.Loader = func(scache * scache.SCache, keys ... string) (interface{}, error) {
		conn, err := qdao.GetDaoManager().Get(scache.Dao);
		if (err != nil) {
			return nil, err;
		}
		var code = keys[0];
		return conn.Get(scache.Db, "", code, true);
	}
	scache_khistory.Loader = func(scache *scache.SCache, keys ... string) (interface{}, error) {
		conn, err := qdao.GetDaoManager().Get(scache.Dao);
		if (err != nil) {
			return nil, err;
		}
		var code = keys[0];
		var datestr = keys[1];
		return conn.Get(scache.Db, code, datestr, true);
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

	defer qerr.SimpleRecover(0);
	var g = global.GetInstance();
	g.Continue = true;
	g.PanicHandler = func(pan interface{}) {
		qlog.Log(qlog.ERROR, pan);
	}
	g.SetData("global", g);
	g.SetData("cachem", scache.GetCacheManager());
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
	var config, err = qconfig.ConfigLoad(g.ConfigPath, "includes")
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
	qref.GetMapperManager().Init(mapperConfig);
	g.SetData("mapperm", qref.GetMapperManager());

	// [redis] ------------------------------------------------------------------------------------------------
	qdao.GetDaoManager().Init();
	g.SetData("daom", qdao.GetDaoManager());

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
