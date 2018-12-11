package httpv

import (
	"github.com/aarzilli/golua/lua"
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/util/util"
	"github.com/camsiabor/qstock/run"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/stevedonovan/luar"
)



func getL() (* lua.State) {
	var L = luar.Init();
	defer L.Close();
	L.OpenLibs();
	var Q = global.GetInstance().Data();
	luar.Register(L, "Q", Q);
	return L;
}

func (o HttpServer) handleLuaCmd(cmd string, m map[string]interface{}, c * gin.Context) {

	var script = util.GetStr(m, "", "script");
	if (len(script) == 0) {
		o.RespJsonEx(0, errors.New("no script content"), c);
		return;
	}

	// TODO arguments
	var L = luar.Init();
	defer L.Close();
	L.OpenLibs();
	var Q = global.GetInstance().Data();
	luar.Register(L, "Q", Q);

	var code = 0;
	var data interface{};
	var err = L.DoString(script);
	if (err == nil) {
		data, err = run.LuaGetVal(L, 1);
		if (err == nil) {
			r2, err := run.LuaGetVal(L, 2);
			if (err == nil && r2 != nil) {
				code = util.AsInt(r2, 0);
			}
		}
	}

	if (err != nil) {
		code = 500;
		var luastacks = L.StackTrace();
		var r = make(map[string]interface{});
		if (luastacks != nil && len(luastacks) > 0) {
			r["func"] = luastacks[0].Name;
			r["line"] = luastacks[0].CurrentLine;
			r["file"] = luastacks[0].ShortSource;
		}
		r["stack"] = luastacks;
		r["err"] = err.Error();
		data = r;
	}
	o.RespJson(code, data, c);
}


