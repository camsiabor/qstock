package main

import (
	"encoding/json"
	"github.com/camsiabor/qcom/agenda"
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/qdao"
	"github.com/camsiabor/qcom/qlog"
	"github.com/camsiabor/qcom/qref"
	"github.com/camsiabor/qcom/qtime"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/dict"
	"github.com/camsiabor/qstock/httpv"
	"github.com/camsiabor/qstock/sync"
	"net"
	"strings"
	"time"
)


func master(g * global.G) {

	var jsonstr, _ = json.Marshal(g.Config)
	qlog.Log(qlog.INFO, "config", string(jsonstr[:]));

	var master_listen = util.GetStr(g.Config, "127.0.0.1:65000", "master", "listen");
	if (!strings.Contains(master_listen, ":")) {
		master_listen = ":" + master_listen;
	}
	var lerr error;
	g.Listener, lerr = net.Listen("tcp", master_listen);
	if (lerr != nil) {
		qlog.Log(qlog.INFO, "establish master listen fail ", master_listen, lerr);
		panic(lerr);
	}
	qlog.Log(qlog.INFO, "master", "listener establish", master_listen);

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

	// [cache] --------------------------------------------------------------------------------------------
	initCacher(g);

	// [api puller] --------------------------------------------------------------------------------------------
	initSyncer(g);


}

func initCacher(g * global.G) {

	var cache_manager = scache.GetCacheManager();
	g.SetData("cachem", cache_manager);

	var cache_timestamp = cache_manager.Get(dict.CACHE_TIMESTAMP);
	cache_timestamp.Loader = func(scache *scache.SCache, factor int, timeout time.Duration, keys ...string) (v interface{}, err error) {
		var key = keys[0];
		if (strings.Contains(key, "@")) {
			return qtime.Time2Int64(nil), nil;
		}
		return time.Now().Format("20060102150405"), nil;
	}


	var scache_code = cache_manager.Get(dict.CACHE_STOCK_CODE);
	scache_code.Dao = dict.DAO_MAIN;
	scache_code.Db = dict.DB_DEFAULT;
	scache_code.Loader = func(scache *scache.SCache, factor int, timeout time.Duration, keys ...string) (interface{}, error) {
		var dao, err = qdao.GetDaoManager().Get(dict.DAO_MAIN);
		if (err != nil) {
			qlog.Error(0, err);
			go func() {
				time.Sleep(time.Duration(10) * time.Second);
				scache.Loader(scache, factor, timeout, keys...);
			}();
			return nil, err;
		}
		var codes, _ = dao.Keys(dict.DB_DEFAULT, "", "*");
		var sz, szn = make([]string, 5000), 0;
		var sh, shn = make([]string, 5000), 0;
		var su, sun = make([]string, 5000), 0;
		var all, alln = make([]string, 15000), 0;
		for _, code := range codes {
			var include = true;
			switch code[0] {
			case '0':
				sz[szn] = code;
				szn++;
			case '3':
				su[szn] = code;
				sun++;
			case '6':
				sh[shn] = code;
				shn++;
			default:
				include = false;
			}
			if (include) {
				all[alln] = code;
				alln++;
			}
		}

		var sz_sh = make([]string, 0);
		sz_sh = append(sz_sh, sz[:szn]...);
		sz_sh = append(sz_sh, sh[:shn]...);

		scache.Set(all[:alln], "all");
		scache.Set(sz[:szn], dict.SHENZHEN);
		scache.Set(sh[:shn], dict.SHANGHAI);
		scache.Set(su[:sun], dict.STARTUP);
		scache.Set(sz_sh, dict.SHENZHEN + "." + dict.SHANGHAI);
		qlog.Log(qlog.INFO, "cache", "code", "shenzhen", szn, "shanghai", shn, "startup", sun, "all", alln);
		return scache, nil;
	}
	scache_code.Loader(scache_code, 1, 0);

	var scache_snapshot = scache.GetCacheManager().Get(dict.CACHE_STOCK_SNAPSHOT);
	scache_snapshot.Dao = dict.DAO_MAIN;
	scache_snapshot.Db = dict.DB_DEFAULT;
	scache_snapshot.Loader = func(scache * scache.SCache, factor int, timeout time.Duration,  keys ... string) (interface{}, error) {
		conn, err := qdao.GetDaoManager().Get(scache.Dao);
		if (err != nil) {
			return nil, err;
		}
		var code = keys[0];
		return conn.Get(scache.Db, "", code, true);
	}


	var scache_khistory = scache.GetCacheManager().Get(dict.CACHE_STOCK_KHISTORY);
	scache_khistory.Dao = dict.DAO_MAIN;
	scache_khistory.Db = dict.DB_HISTORY;
	scache_khistory.Timeout = -1; //time.Second * time.Duration(20);
	scache_khistory.Loader = func(scache *scache.SCache, factor int, timeout time.Duration, keys ... string) (interface{}, error) {
		conn, err := qdao.GetDaoManager().Get(scache.Dao);
		if (err != nil) {
			return nil, err;
		}
		var code = keys[0];
		var datestr = keys[1];
		data, err := conn.Get(scache.Db, code, datestr, true);
		if (data != nil) {
			return data, err;
		}
		if (factor <= 0) {
			return data, err;
		}
		var g = global.GetInstance();
		var cmd = &global.Cmd {
			Service: dict.SERVICE_SYNC,
			Function: "k.history.sz",
			Data : map[string]interface{} {
				"codes" : []string { code },
				"from" : datestr,
			},
		};
		var reply, _ = g.SendCmd(cmd, timeout);
		if (reply != nil) {
			err = reply.RetErr;
			data = reply.RetVal;
		}
		return data, err;
	};
}

func initSyncer(g * global.G) {
	var api_config = util.GetMap(g.Config, true, "api");
	for apiname, api := range api_config {
		var active = util.GetBool(api, true, "active");
		if (!active) {
			continue;
		}
		var syncer = new(sync.Syncer);
		g.CmdHandlerRegister(dict.SERVICE_SYNC, syncer);
		syncer.Run(apiname);
	}
}