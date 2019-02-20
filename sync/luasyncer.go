package sync

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/camsiabor/golua/luar"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/sync/showSdk/httplib"
	"time"
)

func (o Syncer) Lua_request(
	work *ProfileWork,
	fields []string,
	requestargs map[string]interface{}) (ret []interface{}, err error) {

	if !o.doContinue {
		return
	}
	var profile = work.Profile
	var req = httplib.Post(o.domain)
	var reqm = make(map[string]interface{})
	var api = util.GetStr(profile, "", "api")

	reqm["token"] = o.appsecret
	reqm["api_name"] = api
	if fields != nil && len(fields) > 0 {
		reqm["fields"] = fields
	}
	reqm["params"] = requestargs
	reqbody, err := json.Marshal(reqm)
	if err != nil {
		return
	}
	req.Body(reqbody)
	var timeout = util.GetInt64(profile, 20, "timeout")
	var nice = util.GetInt64(profile, 250, "nice")
	req.SetTimeout(time.Duration(10)*time.Second, time.Duration(timeout)*time.Second)
	if nice > 0 {
		time.Sleep(time.Millisecond * time.Duration(nice))
	}
	httpresp, err := req.DoRequest()
	if err != nil {
		return
	}
	var m map[string]interface{}
	var buffer = new(bytes.Buffer)
	buffer.ReadFrom(httpresp.Body)
	err = json.Unmarshal(buffer.Bytes(), &m)
	if err != nil {
		return nil, err
	}

	var retcode = util.GetInt(m, 0, "code")
	if retcode != 0 {
		var retmsg = util.GetStr(m, "", "msg")
		return nil, errors.New(retmsg)
	}

	var data = util.GetMap(m, false, "data")
	var cols = util.GetStringSlice(data, "fields")
	var rows = util.GetSlice(data, "items")
	var datalen = len(rows)
	if datalen <= 0 {
		return nil, nil
	}
	maps, err := util.ColRowToMaps(cols, rows)
	if err != nil {
		return nil, err
	}
	_, _, err = o.PersistAndCache(work, maps)
	return maps, err
}

func (o *Syncer) Lua_handler(phrase string, work *ProfileWork) (interface{}, error) {
	if phrase != "work" {
		return nil, nil
	}

	var L = luar.Init()

	L.Close()
	// TODO
	return nil, nil
}
