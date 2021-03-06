package httpv

import (
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/run/rlua"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func (o *HttpServer) ResolvePath(path string, category string) (string, error) {
	var err error
	category = strings.ToLower(category)
	if category == "lua" {
		path = rlua.GetLuaPath() + path
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return path, err
	}
	return path, err
}

func (o *HttpServer) routeOS() {

	var os_file = o.engine.Group("/os/file")
	os_file.POST("/list", func(c *gin.Context) {

		var m, _ = o.ReqParse(c)
		var path = util.GetStr(m, "", "path")
		var category = util.GetStr(m, "lua", "category")

		path, perr := o.ResolvePath(path, category)
		if perr != nil {
			o.RespJsonEx(nil, perr, c)
			return
		}

		var filterlen int
		var filters []string
		var filter = util.GetStr(m, "", "filter")
		if len(filter) > 0 {
			filters = strings.Split(filter, ",")
			filterlen = len(filters)
		}

		var files, err = ioutil.ReadDir(path)
		if err != nil {
			o.RespJsonEx(nil, err, c)
		}
		var count = len(files)
		var data = make([]interface{}, count)
		for i, file := range files {

			var name = file.Name()
			if filterlen > 0 {
				for i := 0; i < filterlen; i++ {
					if !strings.Contains(name, filters[i]) {
						continue
					}
				}
			}

			var one = make(map[string]interface{})
			one["isdir"] = file.IsDir()
			one["size"] = file.Size()
			one["mode"] = file.Mode().String()
			one["mtime"] = file.ModTime().Second()
			one["name"] = name
			data[i] = one
		}
		o.RespJsonEx(data, err, c)
	})

	os_file.POST("/text", func(c *gin.Context) {
		var data interface{}
		var m, _ = o.ReqParse(c)
		var path = util.GetStr(m, "", "path")
		var category = util.GetStr(m, "", "category")
		path, perr := o.ResolvePath(path, category)
		if perr != nil {
			o.RespJsonEx(nil, perr, c)
			return
		}

		var bytes, err = ioutil.ReadFile(path)
		if err == nil {
			data = string(bytes[:])
		}
		o.RespJsonEx(data, err, c)
	})

	os_file.POST("/write", func(c *gin.Context) {
		var m, _ = o.ReqParse(c)
		var path = util.GetStr(m, "", "path")
		var category = util.GetStr(m, "", "category")
		var text = util.GetStr(m, "", "text")
		path, perr := o.ResolvePath(path, category)
		if perr != nil {
			o.RespJsonEx(nil, perr, c)
			return
		}
		var stat, err = os.Stat(path)
		if err == nil {
			err = ioutil.WriteFile(path, []byte(text), stat.Mode())
		}
		o.RespJsonEx("written", err, c)
	})

	os_file.POST("/delete", func(c *gin.Context) {
		var m, _ = o.ReqParse(c)
		var path = util.GetStr(m, "", "path")
		var category = util.GetStr(m, "", "category")
		path, perr := o.ResolvePath(path, category)
		if perr != nil {
			o.RespJsonEx(nil, perr, c)
			return
		}
		var err = os.Remove(path)
		o.RespJsonEx("deleted", err, c)
	})

}
