package httpv

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/qdao"
	"github.com/camsiabor/qcom/qerr"
	"github.com/camsiabor/qcom/qlog"
	"github.com/camsiabor/qcom/qos"
	"github.com/camsiabor/qcom/qref"
	"github.com/camsiabor/qcom/qtime"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/dict"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// gin doc
// https://godoc.org/github.com/gin-gonic/gin#Context.AbortWithError

type HttpServer struct {
	Root    string
	Rootabs string
	Engine  *gin.Engine
	lock    sync.RWMutex
	data    map[string]interface{}
}

var _instance *HttpServer = &HttpServer{
	data: map[string]interface{}{},
}

func GetInstance() *HttpServer {
	return _instance
}

func (o *HttpServer) GetData(key string) interface{} {
	o.lock.RLock()
	defer o.lock.RUnlock()
	return o.data[key]
}

func (o *HttpServer) SetData(key string, val interface{}) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.data[key] = val
}

func (o *HttpServer) RespJsonEx(data interface{}, err error, c *gin.Context) {
	if err == nil {
		o.RespJson(0, data, c)
	} else {
		o.RespJson(500, err.Error(), c)
	}
}

func (o *HttpServer) RespJson(code int, data interface{}, c *gin.Context) {
	var jstr, ok = data.(string)
	if ok {
		var slen = len(jstr)
		var first = jstr[0]
		var last = jstr[slen-1]
		if (first == '[' && last == ']') || (first == '{' && last == '}') {
			json.Unmarshal([]byte(jstr), &data)
		}
	}
	var fr = map[string]interface{}{
		"code": code,
		"data": data,
	}
	if c.GetBool("zlib") {

		var frbytes, err = json.Marshal(fr)
		if err != nil {
			panic(err)
		}
		var buffer bytes.Buffer
		zlibwriter, _ := zlib.NewWriterLevel(&buffer, zlib.DefaultCompression)
		zlibwriter.Write(frbytes[:])
		zlibwriter.Close()
		//c.Stream(func(w io.Writer) bool {
		//	w.Write(buffer.Bytes())
		//	return false
		//})
		//zlib.NewWriterLevel(&buffer, zlib.DefaultCompression)
		var base64str = base64.StdEncoding.EncodeToString(buffer.Bytes()[:])
		c.String(200, base64str)
	} else {
		c.JSON(200, fr)
	}

}

func (o *HttpServer) ReqParse(c *gin.Context) (map[string]interface{}, error) {
	var bytes, err = ioutil.ReadAll(c.Request.Body)
	if err != nil {
		o.RespJsonEx(nil, err, c)
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		o.RespJsonEx(nil, err, c)
	}
	var zlib = util.GetBool(m, false, "zlib")
	if zlib {
		c.Set("zlib", true)
	}
	return m, err
}

func (o *HttpServer) handleDBCmd(cmd string, m map[string]interface{}, c *gin.Context) {
	var daoname = util.GetStr(m, dict.DAO_MAIN, "dao")
	var dao, err = qdao.GetManager().Get(daoname)
	if err != nil {
		o.RespJsonEx(nil, err, c)
		return
	}
	//var db = util.GetStr(m, pers.DB_DEFAULT, "db");
	var args = util.GetSlice(m, "args")
	var rvals, rerr = qref.FuncCallByName(dao, cmd, args...)
	if rerr != nil {
		o.RespJsonEx(nil, rerr, c)
		return
	}
	var rets = qref.ReflectValuesToList(rvals)
	var retslen = len(rets)
	if retslen > 0 {
		for i := 0; i < retslen; i++ {
			var serr, ok = rets[i].(error)
			if ok {
				o.RespJsonEx(rets, serr, c)
				return
			}
		}
		o.RespJsonEx(rets[0], nil, c)
	} else {
		o.RespJsonEx("success", nil, c)
	}
}

func (o *HttpServer) handleRedisCmd(cmd string, m map[string]interface{}, c *gin.Context) {

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

func (o *HttpServer) handleOSCmd(cmd string, m map[string]interface{}, c *gin.Context) {
	var args = util.GetSlice(m, "args")
	var sargs []string
	if args == nil {
		sargs = make([]string, 0)
	} else {
		sargs = make([]string, len(args))
		for index, one := range args {
			var sone = util.AsStr(one, "")
			sargs[index] = sone
		}
	}
	var timeout = util.GetInt(m, 15, "timeout")
	if timeout <= 0 {
		timeout = 1
	}

	stdoutstr, stderrstr, dotimeout, err := qos.ExecCmd(timeout, cmd, sargs...)
	if err != nil {
		o.RespJsonEx(nil, err, c)
		return
	}
	o.RespJson(0, map[string]interface{}{
		"stdout":  stdoutstr,
		"stderr":  stderrstr,
		"timeout": dotimeout,
	}, c)
}

func (o *HttpServer) handlePanicCmd(c *gin.Context, cmdtype string, cmd string, m map[string]interface{}) {
	var err = util.AsError(recover())
	if err == nil {
		panic(err)
		return
	}
	var info = qref.StackInfo(3)
	info["err"] = err.Error()
	info["a.cmdtype"] = cmdtype
	info["a.cmd"] = cmd
	info["a.m"] = m
	o.RespJson(500, info, c)
}

func (o *HttpServer) handleTimeCmd(cmd string, m map[string]interface{}, c *gin.Context) {

	var key = util.GetStr(m, "js", "key")
	var cache = scache.GetManager().Get("timestamp")
	cmd = strings.ToLower(cmd)
	switch cmd {
	case "now":
		var val = qtime.Time2Int64(nil)
		o.RespJsonEx(val, nil, c)
	case "get":
		var val, err = cache.Get(true, key)
		o.RespJsonEx(val, err, c)
	case "set":
		var val = util.Get(m, nil, "val")
		if val == nil {
			val = time.Now().Format("20060102150405")
		}
		cache.Set(val, key)
		o.RespJson(0, key+" set", c)
	}
}

func (o *HttpServer) routeCmd() {
	var group = o.Engine.Group("/cmd")
	group.GET("/ping", func(c *gin.Context) {
		o.RespJson(0, "pong", c)
	})

	group.POST("/go", func(c *gin.Context) {
		var m, _ = o.ReqParse(c)
		var cmd = util.GetStr(m, "", "cmd")
		if len(cmd) == 0 {
			o.RespJson(500, "give me a command", c)
			return
		}
		var cmdtype = util.GetStr(m, "db", "type")
		defer o.handlePanicCmd(c, cmdtype, cmd, m)
		switch cmdtype {
		case "db":
			o.handleDBCmd(cmd, m, c)
		case "redis":
			o.handleRedisCmd(cmd, m, c)
		case "lua":
			o.handleLuaCmd(cmd, m, c)
		case "os":
			o.handleOSCmd(cmd, m, c)
		case "time":
			o.handleTimeCmd(cmd, m, c)
		}
	})

	var include string
	var g = global.GetInstance()
	var includefile = util.GetStr(g.Config, "res/include.lua", "http", "script", "include")
	var fcontent, ferr = ioutil.ReadFile(o.Rootabs + "/" + includefile)
	if ferr == nil {
		include = string(fcontent[:]) + "\n"
	} else {
		include = ""
	}
	o.data["include"] = include
	group.POST("/query", func(c *gin.Context) {
		var m, _ = o.ReqParse(c)
		var script = util.GetStr(m, "", "script")
		var values = util.GetSlice(m, "values")
		var include = o.data["include"].(string)
		var dao, _ = qdao.GetManager().Get(dict.DAO_MAIN)
		var data, err = dao.Script(dict.DB_DEFAULT, "", "", include+script, values, nil)
		o.RespJsonEx(data, err, c)
	})
}

func (o *HttpServer) routeScript() {
	var group = o.Engine.Group("/script")

	group.POST("/update", func(c *gin.Context) {
		var m, _ = o.ReqParse(c)
		var name = util.GetStr(m, "", "name")
		var dao, _ = qdao.GetManager().Get(dict.DAO_MAIN)
		var data, err = dao.Update(dict.DB_COMMON, "script", name, m, true, -1, nil)
		o.RespJsonEx(data, err, c)
	})

	group.POST("/list", func(c *gin.Context) {
		var dao, _ = qdao.GetManager().Get(dict.DAO_MAIN)
		var scripts, err = dao.Keys(dict.DB_COMMON, "script", "*", nil)
		o.RespJsonEx(scripts, err, c)
	})

	group.POST("/get", func(c *gin.Context) {
		var m, _ = o.ReqParse(c)
		var name = util.GetStr(m, "", "name")
		var dao, _ = qdao.GetManager().Get(dict.DAO_MAIN)
		var data, err = dao.Get(dict.DB_COMMON, "script", name, 1, nil)
		o.RespJsonEx(data, err, c)
	})

	group.POST("/delete", func(c *gin.Context) {
		var m, _ = o.ReqParse(c)
		var name = util.GetStr(m, "", "name")
		var dao, _ = qdao.GetManager().Get(dict.DAO_MAIN)
		var data, err = dao.Delete(dict.DB_COMMON, "script", name, nil)
		o.RespJsonEx(data, err, c)
	})
}

func (o *HttpServer) routeStatic() {

	//_router.LoadHTMLGlob(_rootabs + "/page/*")

	var router = o.Engine
	var rootabs = o.Rootabs

	router.Static("/js", rootabs+"/js")
	router.Static("/css", rootabs+"/css")
	router.Static("/img", rootabs+"/img")
	router.Static("/res", rootabs+"/res")
	router.Static("/svg", rootabs+"/svg")
	router.Static("/tmp", rootabs+"/tmp")
	router.Static("/h", rootabs+"/page")

	router.HTMLRender = CustomHTMLRenderer{}

	router.GET("/v", func(c *gin.Context) {
		var page = c.Query("p")
		c.HTML(200, page, &o)
	})
}

func (o *HttpServer) Run() {

	o.data = make(map[string]interface{})

	var g = global.GetInstance()
	var config_http = util.GetMap(g.Config, true, "http")
	var active = util.GetBool(config_http, true, "active")
	if !active {
		qlog.Log(qlog.INFO, "http", "not active")
		return
	}

	var err error

	var port = util.GetStr(config_http, "8080", "port")
	o.Root = util.GetStr(config_http, "../web", "root")
	o.Rootabs, err = filepath.Abs(o.Root)
	if err != nil {
		qlog.Log(qlog.ERROR, "http", "root", err)
		return
	}
	qlog.Log(qlog.INFO, "http", "port", port, "root", o.Root)

	var logfilepath = util.GetStr(config_http, "log/http.log", "log", "file")
	if !strings.Contains(logfilepath, "console") {
		var logfile, err = os.OpenFile(logfilepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err == nil {
			qlog.Log(qlog.INFO, "http", "log", logfilepath)
		} else {
			qlog.Log(qlog.ERROR, "http", "log", logfilepath, err)
		}
		gin.DefaultWriter = io.MultiWriter(logfile)
		gin.DefaultErrorWriter = io.MultiWriter(logfile)
	}
	var mode = util.GetStr(config_http, gin.ReleaseMode, "mode")
	mode = strings.ToLower(mode)
	gin.SetMode(mode)
	qlog.Log(qlog.INFO, "http", "mode", mode)

	var cache_timestamp = scache.GetManager().Get(dict.CACHE_TIMESTAMP)
	o.Engine = gin.Default()
	o.Engine.Use(func(c *gin.Context) {
		if c.Request.Method == "GET" {
			var path = c.Request.URL.Path
			if path[len(path)-1] == 'l' { // html, last char is l
				var v, _ = cache_timestamp.Get(true, "js")
				c.SetCookie("_u_js", v.(string), 0, "/", "/", false, false)
			}
		}
	})
	o.Engine.Use(Recovery(func(c *gin.Context, err interface{}) {
		var info = qref.StackInfo(2)
		info["err"] = err
		o.RespJson(500, info, c)
	}))

	o.routeStatic()
	o.routeCmd()
	o.routeStock()
	o.routeScript()

	//var refresh_interval = util.GetInt(config_http, 300, "refresh_interval");
	//go GinRefreshPage(refresh_interval);
	go func() {
		defer qerr.SimpleRecover(0)
		qlog.Log(qlog.INFO, "http", "ready to run")
		var err = o.Engine.Run(":" + port) // listen and serve on 0.0.0.0:8080
		if err != nil {
			qlog.Log(qlog.ERROR, "http", "run", err)
		}
	}()
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
