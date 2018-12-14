package sync

import (
	"github.com/camsiabor/qcom/agenda"
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/qdao"
	"github.com/camsiabor/qcom/qerr"
	"github.com/camsiabor/qcom/qlog"
	"github.com/camsiabor/qcom/qref"
	"github.com/camsiabor/qcom/qtime"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/dict"
	"github.com/pkg/errors"
	"strings"
	"sync"
	"time"
)


const SHOWAPI_FETCH_META_TOKEN = "meta.";

type Syncer struct {
	Name               string;
	appid              int;
	domain             string;
	appsecret          string;
	doContinue         bool;
	concurrent         int64;
	mutex              sync.Mutex;
	channelFetchCmd    chan * global.Cmd;
	channelWorkProfile chan * global.Cmd;
	profileRunInfos    map[string]*ProfileRunInfo;
}

type ProfileWork struct {
	Id interface{};
	Dao qdao.D;
	ProfileName string;
	Profile map[string]interface{};
	StartTime int64;
	EndTime int64;
	Force bool;
	Factor float64;
	Context * Syncer;
	GCmd * global.Cmd;
	Args map[string]interface{};
}

func (o * ProfileWork) GetDao() (dao qdao.D, err error) {
	if (o.Dao == nil) {
		if (o.Profile != nil) {
			daoname := util.GetStr(o.Profile, dict.DAO_MAIN, "dao");
			o.Dao, err = qdao.GetDaoManager().Get(daoname);
		}
	}
	return o.Dao, err;
}


type ProfileRunInfo struct {
	RunCount int;
	LastRunTime int64;
	LastEndTime int64;
	LastRunError error;
}


func (o * Syncer) Run(name string) {
	o.Name = name;
	var g = global.GetInstance();
	var config = util.GetMap(g.Config, true, "api", o.Name);
	o.appid =  util.GetInt(config,  0, "appid");
	o.appsecret = util.GetStr(config,  "", "appsecret");
	o.domain = util.GetStr(config, "", "domain");
	o.channelFetchCmd = make(chan * global.Cmd, 64);
	o.channelWorkProfile = make(chan * global.Cmd, 64);
	o.profileRunInfos = make(map[string]*ProfileRunInfo);
	o.doContinue = true;
	o.concurrent = util.GetInt64(g.Config, 1, "api", o.Name, "concurrent");

	var profiles = util.GetMap(config, true, "profiles");
	for _, one := range profiles {
		var profile = util.AsMap(one, false);
		if (profile == nil) {
			continue;
		}
		util.MapMerge(profile, config, false);
	}

	for _, one := range profiles {
		var profile = util.AsMap(one, false);
		if (profile == nil) {
			continue;
		}
		var embed_name = util.GetStr(profile, "", "embed");
		if (len(embed_name) == 0) {
			continue;
		}
		var embed_profile =  util.GetMap(profiles, false, embed_name);
		if (embed_profile == nil) {
			continue;
		}
		util.MapMerge(profile, embed_profile, false);
	}

	var i int64;
	for i = 0; i < o.concurrent; i++ {
		go o.worker();
	}
	go o.heartbeat();
}

func (o * Syncer) stop() {
	o.doContinue = false;
	if (o.channelFetchCmd != nil) {
		close(o.channelFetchCmd);
	}
	if (o.channelWorkProfile != nil) {
		close(o.channelWorkProfile);
	}
}

func (o * Syncer) HandleCmd(cmd * global.Cmd) (interface{}, error) {
	var profileName = cmd.Function;
	if (len(profileName) == 0) {
		return nil, errors.New("profile name is null");
	}

	o.channelFetchCmd <- cmd;
	return cmd.RetVal, cmd.RetErr;
}


func (o * Syncer) heartbeat() {
	var g = global.GetInstance();
	var select_interval = 1;
	for {
		var timeout = time.After(time.Duration(select_interval) * time.Second)

		var ok bool;
		var cmd * global.Cmd;
		select {
			case cmd, ok = <- o.channelFetchCmd:
				if (!ok) {
					qlog.Log(qlog.INFO, o.Name, "receive channel close");
				}
			case <-timeout:
				cmd = nil;
		}
		if (!o.doContinue) {
			break;
		}
		//qlog.Log(qlog.INFO, o.Name, "heartbeat cmd", cmd);
		var agendaNameDefault = util.GetStr(g.Config, "stock", "api", o.Name, "agenda");
		var profiles = util.GetMap(g.Config, true, "api", o.Name, "profiles");
		for profilename, profile := range profiles {
			if (util.AsMap(profile, false) == nil) {
				continue;
			}
			var force = false;
			if (cmd != nil) {
				if (strings.Contains(cmd.Function, profilename)) {
					force = strings.Contains(cmd.Cmd, "force");
				} else {
					continue;
				}
			}
			var factor = 1.0;
			if (!force) {
				var agendaName = util.GetStr(profile, agendaNameDefault, "agenda");
				var agendi = agenda.GetAgendaManager().Get(agendaName);
				if (agendi != nil) {
					var slice = agendi.In(nil);
					if (slice == nil) {
						continue;
					}
					factor = util.GetFloat64(slice, 1, "factor");
				}
				var interval = util.GetInt64(profile, 0, "interval");
				if (interval <= 0 || factor <= 0) {
					continue;
				}
			}
			var dcmd * global.Cmd;
			if (cmd == nil) {
				dcmd = &global.Cmd{ Function : profilename }
			} else {
				dcmd = cmd;
			}
			dcmd.SetData("factor", factor);
			o.channelWorkProfile <- dcmd;
		}
		select_interval = util.GetInt(g.Config, 300, "api", o.Name, "select_interval");
	}
}

func (o * Syncer) worker() {
	qlog.Log(qlog.INFO, "api", o.Name, "worker start");
	for cmd := range o.channelWorkProfile {
		if (!o.doContinue) {
			break;
		}
		var profilename = cmd.Function;
		factor := util.AsFloat64(cmd.GetData("factor"), 1);
		force := strings.Contains(cmd.Cmd, "force");
		qlog.Log(qlog.DEBUG, o.Name, "worker", "receive profilename", profilename);
		var profile = o.GetProfile(profilename);
		if (profile == nil) {
			qlog.Log(qlog.ERROR, o.Name, "worker", "profile not found", profilename);
		} else {
			var work = &ProfileWork{
				Profile: profile,
				ProfileName: profilename,
				Force: force,
				Factor: factor,
				GCmd : cmd,
			};
			o.DoProfileWithRecord(work)
		}
		if (!o.doContinue) {
			break;
		}
	}
	qlog.Log(qlog.INFO, "api", "worker end", o.concurrent);
}

func (o * Syncer) DoProfileRecover(work * ProfileWork) {
	var err = recover();
	if (err == nil) {
		return;
	}
	if (work != nil && work.GCmd != nil && work.GCmd.RetChan != nil) {
		work.GCmd.RetErr = util.AsError(err);
		work.GCmd.RetChan <- work.GCmd;
	}
}

func (o * Syncer) GetProfile(name string) map[string]interface{} {
	var g  = global.GetInstance();
	return util.GetMap(g.Config, false, "api", o.Name, "profiles", name);
}

func (o * Syncer) DoProfile(work * ProfileWork) ([]interface{}, error) {

	defer o.DoProfileRecover(work);

	if (work.Dao == nil) {
		work.Dao, _ = work.GetDao();
	}

	if (work.Profile == nil) {
		work.Profile = o.GetProfile(work.ProfileName);
	}

	var now = time.Now();
	work.StartTime = now.Unix();
	work.Id = qtime.Time2Int64(&now);

	var retvalwrap []interface{};
	var funcname = util.GetStr(work.Profile, "", "handler");
	var retvals, err = qref.FuncCallByName(o, funcname, "work", work);
	if (retvals != nil) {
		retvalwrap = make([]interface{}, len(retvals));
		for i, retval := range retvals {
			if (retval.IsValid()) {
				retvalwrap[i] = retval.Interface();
			}
		}
	}

	work.EndTime = time.Now().Unix();

	if (work.GCmd != nil) {
		if (work.GCmd.RetChan != nil) {
			work.GCmd.RetErr = err;
			work.GCmd.RetVal = retvals;
			work.GCmd.RetChan <- work.GCmd;
		}
	}
	return retvalwrap, err;
}

func (o * Syncer) DoProfileWithRecord(work * ProfileWork) (ferr error) {

	defer qerr.SimpleRecover(0);

	work.Context = o;
	var profile = work.Profile;
	var profilename = work.ProfileName;

	var now = time.Now();
	var profileRunInfo = o.GetProfileRunInfo(profilename);
	var interval = util.GetInt64(profile, 3600, "interval");
	interval = int64(float64(interval) * work.Factor);
	var colddown = interval / 10;
	if (!work.Force) {
		if (now.Unix() - profileRunInfo.LastRunTime < colddown) {
			//qlog.Log(qlog.INFO, o.Name, "current running", profilename, profileRunInfo.LastRunTime, "/", now.Unix());
			return nil;
		}
	}
	profileRunInfo.LastRunTime = now.Unix();
	database := util.GetStr(profile, dict.DB_DEFAULT, "db");

	dao, cerr := work.GetDao();
	if (dao == nil) {
		return cerr;
	}

	start := now;
	timestamp := now.Unix();

	metatoken := o.GetMetaToken(profilename);

	slast, _ := dao.Get(database, metatoken, "last", false);
	last := util.AsInt64(slast, 0);

	if (work.Force) {
		qlog.Log(qlog.INFO, o.Name, profilename, "force!", "current", timestamp, "last", last);
	} else {
		if (timestamp - last < interval) {
			return nil;
		}
	}
	qlog.Log(qlog.INFO, o.Name, profilename, "current", timestamp, "last", last, "interval", interval, "delta", timestamp -last, "factor", work.Factor);

	var sync_record_cacher = scache.GetCacheManager().Get("sync");

	dao.Update(database, metatoken, "start", timestamp, true, -1);
	dao.Update(database, metatoken, "start_id", work.Id, true, -1);
	dao.Update(database, metatoken, "start_str", qtime.YYYY_MM_dd_HH_mm_ss(&start), true, -1);

	sync_record_cacher.SetSubVal(timestamp, profilename, "start");
	sync_record_cacher.SetSubVal(qtime.YYYY_MM_dd_HH_mm_ss(&start), profilename, "start_str");

	var end = time.Now();
	var elapse = end.Unix() - start.Unix();
	var data, err = o.DoProfile(work);
	var count int;
	if (data == nil) {
		count = 0;
	} else {
		count = len(data);
	}

	if (err == nil) {
		start = time.Now();
		profile["last"] = start.Unix();
		dao.Update(database, metatoken, "last", start.Unix(), true, -1);
		dao.Update(database, metatoken, "last_id", work.Id, true, -1);
		dao.Update(database, metatoken, "last_str", qtime.YYYY_MM_dd_HH_mm_ss(&start), true, -1);
		dao.Update(database, metatoken, "last_count", count, true, -1);

		profileRunInfo.LastEndTime = end.Unix();
		profileRunInfo.RunCount = profileRunInfo.RunCount + 1;

		sync_record_cacher.SetSubVal(work.EndTime, profilename, "last");
		sync_record_cacher.SetSubVal(work.Id, profilename, "last_id");
		sync_record_cacher.SetSubVal(qtime.YYYY_MM_dd_HH_mm_ss(&end), profilename, "last_str");

		qlog.Log(qlog.INFO, "profile", profilename, "done", "consume", elapse);

	} else {
		profileRunInfo.LastRunError = err;
		qlog.Log(qlog.ERROR, "profile", profilename, err.Error(), "consume", elapse);
	}

	profileRunInfo.LastRunTime = 0;

	return err;
}

func (o * Syncer) GetMetaToken(profilename string) (string) {
	return SHOWAPI_FETCH_META_TOKEN + profilename;
}



func (o * Syncer) GetProfileRunInfo(profilename string) (*ProfileRunInfo) {
	var runinfo = o.profileRunInfos[profilename];
	if (runinfo == nil) {
		o.mutex.Lock();
		runinfo = o.profileRunInfos[profilename];
		if (runinfo == nil) {
			runinfo = new(ProfileRunInfo);
			runinfo.RunCount = 0;
			runinfo.LastEndTime = 0;
			runinfo.LastRunTime = 0;
			o.profileRunInfos[profilename] = runinfo;
		}
		o.mutex.Unlock()
	}
	return runinfo;
}



func (o * Syncer) PersistAndCache(
	work * ProfileWork,
	data []interface{}) (rdata []interface{}, ids []interface{}, err error){

	if (data == nil) {
		return nil, nil, nil;
	}
	var datalen = len(data);
	if (datalen <= 0) {
		return nil, nil, nil;
	}
	rdata = data;
	var profile = work.Profile;
	var key = util.GetStr(profile, "code", "key");
	var group = util.GetStr(profile, "", "group");
	var groupkey = util.GetStr(profile, "", "groupkey");

	var mappername = util.GetStr(profile, "", "mapper");
	var mapper = qref.GetMapperManager().Get(mappername);

	var idsss = make([]string, datalen);
	ids = make([]interface{}, datalen);
	for i, one := range data {
		var m = one.(map[string]interface{});
		m["_u"] = work.Id;
		if (mapper != nil) {
			_, err := mapper.Map(one, false);
			if (err != nil) {
				return nil, nil, err;
			}
		}
		idsss[i] = util.GetStr(m, "", key);
		ids[i] = idsss[i];
		if (len(group) == 0 && len(groupkey) > 0) {
			group = util.GetStr(one, "", groupkey);
		}
	}

	var db = util.GetStr(profile, "", "db");
	var cachername = util.GetStr(profile, "", "cacher");
	var cacher = scache.GetCacheManager().Get(cachername);
	_, err = work.Dao.Updates(db, group, ids, data, true, -1);
	if (cacher != nil) {
		if (len(group) == 0) {
			cacher.Sets(data, idsss);
		} else {
			cacher.SetSubVals(data, idsss, group);
		}
	}
	return data, ids, err;
}
