package sync

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/camsiabor/qcom/qdao"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"showSdk/httplib"
	"time"
)

// https://www.shenjian.io/index.php?r=market/product&product_id=328#stack-info-2

func (o Syncer) ShenJian_request(
	dao qdao.D,
	profile map[string]interface{},
	profilename string,
	fields []string,
	requestargs map[string]interface{},
	handler SyncAPIHandler) (data []interface{}, ids []interface{}, err error) {

	if (!o.doContinue) {
		return;
	}

	var api = util.GetStr(profile, "", "api");
	var url = o.domain + "/" + api;
	var req = httplib.Post(url)
	httpresp, err := req.DoRequest();
	if err != nil {
		return;
	}
	var m map[string]interface{};
	var buffer = new(bytes.Buffer);
	buffer.ReadFrom(httpresp.Body);
	err = json.Unmarshal(buffer.Bytes(), &m)
	if (err != nil) {
		return nil, nil, err;
	}

	var retcode = util.GetInt(m, 0, "error_code");
	if (retcode != 0) {
		var retmsg = util.GetStr(m, "", "reason");
		return nil, nil, errors.New(retmsg);
	}

	var key = util.GetStr(profile, "code", "key");
	var mappername = util.GetStr(profile, "", "mapper");
	var mapper = util.GetMapperManager().Get(mappername);


	data = util.GetSlice(m, "data");
	ids = make([]interface{}, len(data));
	var updatetime = time.Now().Format("02-1504"); // updateimte
	for i, onedata := range data {
		var info = onedata.(map[string]interface{});
		info["_u"] = updatetime;
		ids[i] = info[key];
		if (mapper != nil) {
			_, err := mapper.Map(info, false);
			if (err != nil) {
				return nil, nil, err;
			}
		}
	}

	if (err == nil) {
		var db = util.GetStr(profile, "", "db");
		var group = util.GetStr(profile, "", "group");
		var cachername = util.GetStr(profile, "", "cacher");
		var cacher = scache.GetCacheManager().Get(cachername);

		var idsss = util.AsStringSlice(ids, 0);
		if (len(group) == 0) {
			_, err = dao.Updates(db, group, ids, data, true, false);
			if (cacher != nil) {
				cacher.Sets(data, idsss);
			}
		} else {
			_, err = dao.Updates(db, group, ids, data, true, true);
			if (cacher != nil) {
				cacher.SetSubVals(data, idsss, group);
			}
		}
	}
	return data, ids, err;
}


func (o * Syncer) ShenJian_snapshot(
	phrase string, dao qdao.D,
	profile map[string]interface{}, profilename string,
	arg1 interface{}, arg2 interface{} ) (err error) {

	if (phrase != "work") {
		return nil;
	}
	_, _, err = o.ShenJian_request(dao, profile, profilename, nil, nil, nil);
	return err;
}


