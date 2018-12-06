package httpv

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"qcom/global"
	"qcom/qdao"
	"qcom/scache"
	"qcom/util"
	"qcom/util/qlog"
	"qcom/util/qos"
	"qcom/util/qref"
	"qcom/util/qtime"
	"stock/dict"
	"strings"
	"time"
)

// gin doc
// https://godoc.org/github.com/gin-gonic/gin#Context.AbortWithError

type HttpServer struct {
	Root string;
	Rootabs string;
	Engine * gin.Engine;
	Data map[string]interface{};
}

var _instance * HttpServer = &HttpServer{}

func  GetInstance() (* HttpServer) {
	return _instance;
}

func (o * HttpServer) RespJsonEx(data interface{}, err error, c * gin.Context) {
	if (err == nil) {
		o.RespJson(0, data, c);
	} else {
		o.RespJson(500, err.Error(), c);
	}
}

func (o * HttpServer) RespJson(code int, data interface{}, c *gin.Context) {
	var jstr, ok = data.(string);
	if (ok) {
		var slen = len(jstr);
		var first = jstr[0];
		var last = jstr[slen - 1];
		if ((first == '[' && last == ']') || (first == '{' && last == '}' )) {
			json.Unmarshal([]byte(jstr), &data);
		}
	}
	var fr = map[string]interface{} {
		"code" : code,
		"data" : data,
	};
	c.JSON(200, fr);
}

func (o * HttpServer) ReqParse(c * gin.Context) (map[string]interface{}, error) {
	var bytes, err = ioutil.ReadAll(c.Request.Body);
	if (err != nil) {
		o.RespJsonEx(nil, err, c);
		return nil, err;
	}
	var m map[string]interface{};
	err = json.Unmarshal(bytes, &m);
	if (err != nil) {
		o.RespJsonEx(nil, err, c);
	}
	return m, err;
}

func (o * HttpServer) handleDBCmd(cmd string, m map[string]interface{}, c * gin.Context) {
	var daoname = util.GetStr(m, dict.DAO_MAIN, "dao");
	var dao, err = qdao.GetDaoManager().Get(daoname);
	if (err != nil) {
		o.RespJsonEx(nil, err, c);
		return;
	}
	//var db = util.GetStr(m, pers.DB_DEFAULT, "db");
	var args = util.GetSlice(m,  "args");
	var rvals, rerr = qref.FuncCallByName(dao, cmd, args...);
	if (rerr != nil) {
		o.RespJsonEx(nil, rerr, c);
		return;
	}
	var rets = qref.ReflectValuesToList(rvals);
	var retslen = len(rets);
	if (retslen > 0) {
		for i := 0; i < retslen; i++ {
			var serr, ok = rets[i].(error);
			if (ok) {
				o.RespJsonEx(rets, serr, c);
				return;
			}
		}
		o.RespJsonEx(rets[0], nil, c);
	} else {
		o.RespJsonEx("success", nil, c);
	}
}

func (o * HttpServer) handleRedisCmd(cmd string, m map[string]interface{}, c * gin.Context) {

	//var db = util.GetStr(m, pers.DB_DEFAULT, "db");
	//var args = util.GetList(m,  "args");
	//rclient, err := pers.GetDaoManager().Get(db);
	//if (err != nil) {
	//	o.RespJsonEx(nil, err, c);
	//	return;
	//}
	//retval, err := pers.RCmd(rclient, cmd, args...);
	//o.RespJsonEx(nil, nil, c);

}

func (o * HttpServer) handleOSCmd(cmd string, m map[string]interface{}, c * gin.Context) {
	var args = util.GetSlice(m,  "args");
	var sargs []string;
	if (args == nil) {
		sargs = make([]string, 0);
	} else {
		sargs = make([]string, len(args));
		for index, one := range args {
			var sone = util.AsStr(one, "");
			sargs[index] = sone;
		}
	}
	var timeout = util.GetInt(m, 15, "timeout");
	if (timeout <= 0) {
		timeout = 1;
	}

	stdoutstr, stderrstr, dotimeout, err := qos.ExecCmd(timeout, cmd, sargs...);
	if (err != nil) {
		o.RespJsonEx(nil, err, c);
		return;
	}
	o.RespJson(0, map[string]interface{}{
		"stdout" : stdoutstr,
		"stderr" : stderrstr,
		"timeout" : dotimeout,
	}, c);
}

func (o * HttpServer) handlePanicCmd(c * gin.Context, cmdtype string, cmd string, m map[string]interface{}) {
	var err = util.AsError(recover());
	if (err == nil) {
		panic(err);
		return;
	}
	var info = qref.StackInfo(3);
	info["err"] = err.Error();
	info["a.cmdtype"] = cmdtype;
	info["a.cmd"] = cmd;
	info["a.m"] = m;
	o.RespJson(500, info, c);
}

func (o * HttpServer) routeCmd() {
	var group = o.Engine.Group("/cmd");
	group.GET("/ping", func(c *gin.Context) {
		o.RespJson(0, "pong", c);
	});

	group.POST("/go", func(c * gin.Context) {
		var m, _ = o.ReqParse(c);
		var cmd = util.GetStr(m, "", "cmd");
		if (len(cmd) == 0) {
			o.RespJson(500, "give me a command", c);
			return;
		}
		var cmdtype = util.GetStr(m, "db", "type");

		defer o.handlePanicCmd(c, cmdtype, cmd, m);
		switch cmdtype {
		case "db" :
			o.handleDBCmd(cmd, m, c);
		case "redis" :
			o.handleRedisCmd(cmd, m, c);
		case "lua":
			o.handleLuaCmd(cmd, m, c);
		case "os":
			o.handleOSCmd(cmd, m, c);
		}
	});

	var include string;
	var g = global.GetInstance();
	var includefile = util.GetStr(g.Config, "res/include.lua", "http", "script", "include");
	var fcontent, ferr = ioutil.ReadFile(o.Rootabs + "/" + includefile);
	if (ferr == nil) {
		include = string(fcontent[:]) + "\n";
	} else {
		include = "";
	}
	o.Data["include"] = include;
	group.POST("/query", func(c * gin.Context) {
		var m, _ = o.ReqParse(c);
		var script = util.GetStr(m, "", "script");
		var values = util.GetSlice(m,  "values");
		var include = o.Data["include"].(string);
		var dao, _ = qdao.GetDaoManager().Get(dict.DAO_MAIN);
		var data, err = dao.Script(dict.DB_DEFAULT, "", "", include + script, values)
		o.RespJsonEx(data, err, c);
	});
}

func (o * HttpServer) routeQuery() {
	var group = o.Engine.Group("/stock");

	group.POST("/timestamp", func(c * gin.Context) {
		var cache = scache.GetCacheManager().Get("sync");
		var retm, err = cache.GetAll();
		o.RespJsonEx(retm, err, c);
	});

	group.POST("/sync", func(c *gin.Context) {
		var m, _ = o.ReqParse(c);
		var profile = util.GetStr(m, "", "profile");
		var cmd = util.GetStr(m, "force", "cmd");
		var _, err = global.GetInstance().SendCmd(&global.Cmd{
			Service: "sync",
			Name: profile,
			Cmd: cmd,
		});
		o.RespJsonEx("cmd sent", err, c);
	});

	group.POST("/clear", func(c *gin.Context) {
		var m, _ = o.ReqParse(c);
		var db = util.GetStr(m, "", "db");
		var group = util.GetStr(m, "", "group");
		var daoname = util.GetStr(m, dict.DAO_MAIN, "dao");
		dao, err := qdao.GetDaoManager().Get(daoname);
		if (err != nil) {
			o.RespJsonEx(nil, err, c);
			return;
		}
		keys, err := dao.Keys(db, group, "*");
		if (err != nil) {
			o.RespJsonEx(nil, err, c);
			return;
		}
		var ids = util.AsSlice(keys, 0);
		ret, err := dao.Deletes(db, group, ids);
		o.RespJsonEx(ret, err, c);
	})

	group.POST("/keys", func(c * gin.Context) {
		var m, _  = o.ReqParse(c);
		var dbs = util.GetSlice(m, "dbs");
		var group = util.GetStr(m, "", "group");
		var keys = util.GetStr(m, "meta*", "keys");
		var ret = make(map[string]interface{});
		var dao, _ = qdao.GetDaoManager().Get(dict.DAO_MAIN);
		for _, db := range dbs {
			var sdb = db.(string);
			keysret, err := dao.Keys(sdb, group, keys);
			if (err != nil) {
				o.RespJsonEx(nil, err, c);
				return;
			}
			for _, key := range keysret {
				var oneret, oneerr = dao.Get(sdb, group, key, true);
				if (oneerr == nil) {
					ret[key] = oneret;
				} else {
					ret[key] = oneerr.Error();
				}
			}
		}
		o.RespJsonEx(ret, nil, c);
	});

	group.POST("/gets", func(c * gin.Context) {
		var m, _ = o.ReqParse(c);
		var ocodes = util.Get(m, nil, "codes");
		var time_from_str = util.GetStr(m, "", "time_from");
		var time_to_str = util.GetStr(m, "", "time_to");

		var index = 0;
		var codes = ocodes.([]interface{});
		var cacher_stock_snapshot = scache.GetCacheManager().Get(dict.CACHE_STOCK_SNAPSHOT);
		var cacher_stock_khistory = scache.GetCacheManager().Get(dict.CACHE_STOCK_KHISTORY);

		var time_array_len = 0;
		var time_array []string = nil;
		if (len(time_from_str) > 0) {
			time_array = make([]string, 256);
			if (len(time_to_str) <= 0) {
				var now = time.Now()
				time_to_str = qtime.YYYY_MM_dd(&now);
			}
			time_from, perr := time.Parse("2006-01-02", time_from_str);
			if (perr != nil) {
				o.RespJsonEx(nil, perr, c);
				return;
			}
			time_to, perr := time.Parse("2006-01-02", time_to_str);
			if (perr != nil) {
				o.RespJsonEx(nil, perr, c);
				return;
			}

			time_array, perr := qtime.GetTimeFormatIntervalArray(&time_from, &time_to, "2006-01-02", time.Saturday, time.Sunday);
			if (perr != nil) {
				o.RespJsonEx(nil, perr, c);
				return;
			}
			time_array_len = len(time_array);
		}

		var data = make([]interface{}, len(codes));
		for _, code := range codes {
			if (code == nil) {
				continue;
			}
			var scode = code.(string);
			snapshoto, err := cacher_stock_snapshot.Get(scode, true);
			snapshot := util.AsMap(snapshoto, false);
			if (err != nil) {
				o.RespJsonEx(0, err, c);
				return;
			}
			if (time_array_len > 0) {
				var khistory_subcache = cacher_stock_khistory.GetSub(scode);
				if (khistory_subcache != nil) {
					khistory, err := khistory_subcache.List(true, time_array...);
					if (err != nil) {
						o.RespJsonEx(0, err, c);
						return;
					}
					snapshot["khistory"] = khistory;
				}
			}

			if (snapshot != nil) {
				data[index] = snapshot;
				index = index + 1;
			}
		}

		o.RespJson(0, data[:index], c);
	});


}

func (o * HttpServer) routeScript() {
	var group = o.Engine.Group("/script");

	group.POST("/update", func(c * gin.Context) {
		var m, _ = o.ReqParse(c);
		var name = util.GetStr(m, "", "name");
		var dao, _ = qdao.GetDaoManager().Get(dict.DAO_MAIN);
		var data, err = dao.Update(dict.DB_COMMON, "script", name, m, true, true);
		o.RespJsonEx(data, err, c);
	});

	group.POST("/list", func(c * gin.Context) {
		var dao, _ = qdao.GetDaoManager().Get(dict.DAO_MAIN);
		var scripts, err = dao.Keys(dict.DB_COMMON, "script", "*")
		o.RespJsonEx(scripts, err, c);
	});

	group.POST("/get", func(c * gin.Context) {
		var m, _ = o.ReqParse(c);
		var name = util.GetStr(m, "", "name");
		var dao, _ = qdao.GetDaoManager().Get(dict.DAO_MAIN);
		var data, err = dao.Get(dict.DB_COMMON, "script", name, true);
		o.RespJsonEx(data, err, c);
	});

	group.POST("/delete", func(c * gin.Context) {
		var m, _ = o.ReqParse(c);
		var name = util.GetStr(m, "", "name");
		var dao, _ = qdao.GetDaoManager().Get(dict.DAO_MAIN);
		var data, err = dao.Delete(dict.DB_COMMON, "script", name );
		o.RespJsonEx(data, err, c);
	});
}



func (o * HttpServer) routeStatic() {

	//_router.LoadHTMLGlob(_rootabs + "/page/*")

	var router = o.Engine;
	var rootabs = o.Rootabs;

	router.Static("/js", rootabs + "/js")
	router.Static("/css", rootabs+ "/css")
	router.Static("/img", rootabs+ "/img")
	router.Static("/res", rootabs+ "/res")
	router.Static("/svg", rootabs+ "/svg")
	router.Static("/tmp", rootabs+ "/tmp")
	router.Static("/h", rootabs+ "/page")

	router.HTMLRender = CustomHTMLRenderer{};

	router.GET("/v", func(c * gin.Context) {
		var page = c.Query("p");
		c.HTML(200, page, &o );
	});
}

func (o * HttpServer) Run() {

	o.Data = make(map[string]interface{});

	var g = global.GetInstance();
	var config_http = util.GetMap(g.Config, true, "http");
	var active = util.GetBool(config_http, true,  "active");
	if (!active) {
		qlog.Log(qlog.INFO, "http", "not active");
		return;
	}
	go func() {
		var err error;

		var port = util.GetStr(config_http, "8080",  "port");
		o.Root = util.GetStr(config_http, "../web",  "root");
		o.Rootabs, err = filepath.Abs(o.Root);
		if (err != nil) {
			qlog.Log(qlog.ERROR, "http", "root", err);
			return;
		}
		qlog.Log(qlog.INFO, "http", "port", port, "root", o.Root);

		var logfilepath =util.GetStr(config_http, "log/http.log", "log", "file");
		if (!strings.Contains(logfilepath, "console")) {
			var logfile, err = os.OpenFile(logfilepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
			if (err == nil) {
				qlog.Log(qlog.INFO, "http", "log", logfilepath);
			} else {
				qlog.Log(qlog.ERROR, "http", "log", logfilepath, err);
			}
			gin.DefaultWriter =io.MultiWriter(logfile);
			gin.DefaultErrorWriter = io.MultiWriter(logfile);
		}

		o.Engine = gin.Default()
		o.Engine.Use(Recovery(func( c *gin.Context, err interface{}) {
			var info = qref.StackInfo(2);
			info["err"] = err;
			o.RespJson(500, info, c);
		}))

		o.routeStatic();
		o.routeCmd();
		o.routeQuery();
		o.routeScript();

		qlog.Log(qlog.INFO, "http", "ready to run");

		var refresh_interval = util.GetInt(config_http, 300, "refresh_interval");
		go GinRefreshPage(refresh_interval);

		err = o.Engine.Run(":" + port) // listen and serve on 0.0.0.0:8080
		if (err != nil) {
			qlog.Log(qlog.ERROR, "http", "run", err);
		}
	} ();
}


// TODO logger
/*
func (o * HttpServer) Logger() gin.HandlerFunc {
   logClient := logrus.New()

   //禁止logrus的输出
   src, err := os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
   if err!= nil{
      fmt.Println("err", err)
   }
   logClient.Out = src
   logClient.SetLevel(logrus.DebugLevel)
   apiLogPath := "api.log"
   logWriter, err := rotatelogs.New(
      apiLogPath+".%Y-%m-%d-%H-%M.log",
      rotatelogs.WithLinkName(apiLogPath), // 生成软链，指向最新日志文件
      rotatelogs.WithMaxAge(7*24*time.Hour), // 文件最大保存时间
      rotatelogs.WithRotationTime(24*time.Hour), // 日志切割时间间隔
   )
   writeMap := lfshook.WriterMap{
      logrus.InfoLevel:  logWriter,
      logrus.FatalLevel: logWriter,
   }
   lfHook := lfshook.NewHook(writeMap, &logrus.JSONFormatter{})
   logClient.AddHook(lfHook)


   return func (o * HttpServer) (c *gin.Context) {
      // 开始时间
      start := time.Now()
      // 处理请求
      c.Next()
      // 结束时间
      end := time.Now()
      //执行时间
      latency := end.Sub(start)

      path := c.Request.URL.Path

      clientIP := c.ClientIP()
      method := c.Request.Method
      statusCode := c.Writer.Status()
      logClient.Infof("| %3d | %13v | %15s | %s  %s |",
         statusCode,
         latency,
         clientIP,
         method, path,
      )
   }
}
*/