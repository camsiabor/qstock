package httpv

import (
	"github.com/camsiabor/golua/lua"
	"github.com/camsiabor/golua/luar"
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/qref"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/run"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
	"time"
)

func getL() *lua.State {
	var L = luar.Init()
	defer L.Close()
	L.OpenLibs()
	L.OpenDebug()
	L.OpenTable()
	var Q = global.GetInstance().Data()
	luar.Register(L, "Q", Q)
	return L
}

func (o HttpServer) handleLuaCmd(cmd string, m map[string]interface{}, c *gin.Context) {

	var script = util.GetStr(m, "", "script")
	if len(script) == 0 {
		o.RespJsonEx(0, errors.New("no script content"), c)
		return
	}

	var debug = util.GetBool(m, false, "debug")

	// TODO arguments
	var L = luar.Init()
	defer L.Close()
	L.OpenLibs()
	var Q = global.GetInstance().Data()
	luar.Register(L, "Q", Q)

	var start, end int64
	var consume float64

	var code = 0
	var data interface{}
	var goStackInfo map[string]interface{}

	if debug {
		start = time.Now().UnixNano()
	}
	var err = L.DoStringHandle(script, func(L *lua.State, pan interface{}) {
		goStackInfo = qref.StackInfo(5)
		var stackstr = util.AsStr(goStackInfo["stack"], "")
		stackstr = strings.Replace(stackstr, "\t", "  ", -1)
		goStackInfo["stack"] = strings.Split(stackstr, "\n")
	})
	if debug {
		end = time.Now().UnixNano()
		consume = float64((end - start)) / float64(time.Millisecond)
	}

	if err == nil {
		data, err = run.LuaGetVal(L, 1)
		if err == nil {
			r2, err := run.LuaGetVal(L, 2)
			if err == nil && r2 != nil {
				code = util.AsInt(r2, 0)
			}
		}
	}

	if err == nil {
		if debug {
			var wrap = map[string]interface{}{}
			wrap["debug"] = true
			wrap["data"] = data
			wrap["consume"] = consume
			data = wrap
		}
	} else {
		code = 500
		var luastacks = L.StackTrace()
		var luastacksinfo = run.LuaFormatStackToMap(luastacks)
		var r = make(map[string]interface{})

		var luastackslen = len(luastacksinfo)
		if len(luastacksinfo) > 0 {
			r["func"] = luastacksinfo[luastackslen-1]["func"]
			r["line"] = luastacksinfo[luastackslen-1]["line"]
			r["linesrc"] = luastacksinfo[luastackslen-1]["linesrc"]
		}
		if goStackInfo != nil {
			luastacksinfo = append([]map[string]interface{}{goStackInfo}, luastacksinfo...)
		}

		r["stack"] = luastacksinfo
		r["err"] = err.Error()
		r["type"] = "lua"

		if debug {
			r["cosume"] = consume
		}

		data = r
	}
	o.RespJson(code, data, c)
}
