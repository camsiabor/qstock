package sync

import (
	"github.com/camsiabor/qcom/agenda"
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/qdao"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qcom/util/qlog"
	"github.com/camsiabor/qcom/util/qref"
	"github.com/camsiabor/qcom/util/qtime"
	"github.com/camsiabor/qstock/dict"
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
	RequestHandler 	   SyncRequestHandler
}



type ProfileRunInfo struct {
	RunCount int;
	LastRunTime int64;
	LastEndTime int64;
	LastRunError error;
}

type SyncAPIHandler func(
	phrase string,
	dao qdao.D,
	profile map[string]interface{},
	profilename string,
	arg1 interface{},
	arg2 interface{}) (err error);

type  SyncRequestHandler  func(
	dao qdao.D,
	profile map[string]interface{},
	profilename string,
	requestargs map[string]interface{},
	handler SyncAPIHandler) (interface{}, error);

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
	o.concurrent = 0;
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


func (o *Syncer) FilterCmd(cmd *global.Cmd) bool {
	return cmd.Service == "sync";
}

func (o * Syncer) HandleCmd(cmd * global.Cmd) (interface{}, error) {
	o.channelFetchCmd <- cmd;
	return cmd.RetVal, cmd.RetErr;
}


func (o * Syncer) heartbeat() {
	var g = global.GetInstance();
	var select_interval = 1;
	for {
		var timeout = time.After(time.Duration(select_interval) * time.Second)
		var concurrent = util.GetInt64(g.Config, 1, "api", o.Name, "concurrent");
		for (concurrent > o.concurrent) {
			o.concurrent = o.concurrent + 1;
			go o.worker();
		}
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
		var profiles= util.GetMap(g.Config, true, "api", o.Name, "profiles");
		for profilename, profile := range profiles {
			var force = false;
			if (cmd != nil) {
				if (strings.Contains(cmd.Name, profilename)) {
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
					factor = util.GetFloat64(slice, 1, factor);
				}
				var interval = util.GetInt64(profile, 0, "interval");
				if (interval <= 0 || factor <= 0) {
					continue;
				}
			}
			var dcmd * global.Cmd;
			if (cmd == nil) {
				dcmd = &global.Cmd{ Name : profilename }
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
	qlog.Log(qlog.INFO, "api", "worker start", o.concurrent);
	var g = global.GetInstance();
	for {
		var profilename string = "";
		var concurrent = util.GetInt64(g.Config, 1, "api", o.Name, "concurrent");
		if (concurrent < o.concurrent) {
			break;
		}
		var timeout = time.After(time.Duration(60) * time.Second)
		select {
		case cmd, ok := <- o.channelWorkProfile:
			if !ok {
				qlog.Log(qlog.INFO, o.Name, "worker", "channel close");
				break;
			}
			profilename = cmd.Name;
			factor := util.AsFloat64(cmd.GetData("factor"), 1);
			force := strings.Contains(cmd.Cmd, "force");
			qlog.Log(qlog.INFO, o.Name, "worker", "receive profilename", profilename);
			var profile = util.GetMap(g.Config, false, "api", o.Name, "profiles", profilename);
			if (profile == nil) {
				qlog.Log(qlog.ERROR, o.Name, "worker", "profile not found", profilename);
			} else {
				o.doprofile(profilename, profile, force, factor)
			}
		case <-timeout:

		}
		if (!o.doContinue) {
			break;
		}

	}

	o.concurrent = o.concurrent - 1;
	qlog.Log(qlog.INFO, "api", "worker end", o.concurrent);
}


func (o * Syncer) doprofile(profilename string, profile map[string]interface{}, force bool, factor float64) (error) {


	var now = time.Now();
	var profileRunInfo = o.GetProfileRunInfo(profilename);
	var interval = util.GetInt64(profile, 3600, "interval");
	var colddown = util.GetInt64(profile, 600, "colddown");
	interval = int64(float64(interval) * factor);
	colddown = int64(float64(colddown) * factor);
	if (!force) {
		if (now.Unix()-profileRunInfo.LastRunTime < colddown) {
			//qlog.Log(qlog.INFO, o.Name, "current running", profilename, profileRunInfo.LastRunTime, "/", now.Unix());
			return nil;
		}
	}
	profileRunInfo.LastRunTime = now.Unix();

	daoname := util.GetStr(profile, dict.DAO_MAIN, "dao");
	database := util.GetStr(profile, dict.DB_DEFAULT, "db");

	//marker := util.GetStr(profile, "", "marker");
	dao, cerr := qdao.GetDaoManager().Get(daoname);
	if (dao == nil) {
		return cerr;
	}
	start := now;
	timestamp := now.Unix();

	metatoken := o.GetMetaToken(profilename);

	slast, _ := dao.Get(database, metatoken, "last", false);
	last := util.AsInt64(slast, 0);

	if (force) {
		qlog.Log(qlog.INFO, o.Name, profilename, "force!", "current", timestamp, "last", last);
	} else {
		qlog.Log(qlog.INFO, o.Name, profilename, "current", timestamp, "last", last, "interval", interval, "delta", timestamp -last, "factor", factor);
		if (timestamp - last < interval) {
			//qlog.Log(qlog.INFO, o.Name, profilename, "fetch in cooldown");
			return nil;
		}
	}


	var g = global.GetInstance();
	var sync_record_cacher = scache.GetCacheManager().Get("sync");
	var nice_default = util.GetInt64(g.Config, 0, "api", o.Name, "nice");
	var nice_profile = util.GetInt64(profile, nice_default, "nice");
	profile["nice"] = nice_profile;
	dao.Update(database, metatoken, "start", timestamp, true, false);
	dao.Update(database, metatoken, "start_str", qtime.YYYY_MM_dd_HH_mm_ss(&start), true, false);

	sync_record_cacher.SetSubVal(timestamp, profilename, "start");
	sync_record_cacher.SetSubVal(qtime.YYYY_MM_dd_HH_mm_ss(&start), profilename, "start_str");

	var args = make(map[string]interface{});
	var funcname = util.GetStr(profile, "", "handler");
	var _, err = qref.FuncCallByName(o, funcname, "work", dao, profile, profilename, args, args);
	var end = time.Now();
	var elapse = end.Unix() - start.Unix();

	if (err == nil) {
		start = time.Now();
		profile["last"] = start.Unix();
		dao.Update(database, metatoken, "last", start.Unix(), true, false);
		dao.Update(database, metatoken, "last_str", qtime.YYYY_MM_dd_HH_mm_ss(&start), true, false);

		profileRunInfo.LastEndTime = end.Unix();
		profileRunInfo.RunCount = profileRunInfo.RunCount + 1;

		sync_record_cacher.SetSubVal(end.Unix(), profilename, "last");
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


