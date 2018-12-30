package sync

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/camsiabor/qcom/qlog"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/sync/showSdk/httplib"
	"time"
)

// https://www.shenjian.io/index.php?r=market/product&product_id=328#stack-info-2

func (o Syncer) ShenJian_request(
	work *ProfileWork,
	fields []string,
	requestargs map[string]interface{}) (data []interface{}, ids []interface{}, err error) {

	if !o.doContinue {
		return
	}
	var profile = work.Profile
	var api = util.GetStr(profile, "", "api")
	var url = o.domain + "/" + api
	var req = httplib.Post(url)
	var timeout = util.GetInt64(profile, 60, "timeout")
	var nice = util.GetInt64(profile, 200, "nice")
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
		return nil, nil, err
	}

	var retcode = util.GetInt(m, 0, "error_code")
	if retcode != 0 {
		var retmsg = util.GetStr(m, "", "reason")
		return nil, nil, errors.New(retmsg)
	}

	data = util.GetSlice(m, "data")
	return o.PersistAndCache(work, data)
}

func (o *Syncer) ShenJian_snapshot(
	phrase string, work *ProfileWork) (err error) {

	if phrase != "work" {
		return nil
	}
	for i := 0; i < 2; i++ {
		_, _, err = o.ShenJian_request(work, nil, nil)
		if err == nil {
			qlog.Log(qlog.INFO, work.ProfileName, "persist", "success")
			break
		} else {
			qlog.Log(qlog.ERROR, work.ProfileName, "persist", "fail", err)
		}
	}
	return err
}
