package sync

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/camsiabor/qcom/qdao"
	"github.com/camsiabor/qcom/util/util"
	"showSdk/httplib"
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

	data = util.GetSlice(m, "data");
	return o.PersistAndCache(profile, dao, data);
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


