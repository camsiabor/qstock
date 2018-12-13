package main


/*
// TODO daemon process
// TODO actor, proactor
// TODO micro service
// TODO distribute
// TODO elasticsearch mongodb
 */

import (
	"flag"
	"fmt"
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/qconfig"
	"github.com/camsiabor/qcom/qerr"
	"github.com/camsiabor/qcom/qlog"
	"github.com/camsiabor/qcom/qos"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/wrap"
	"github.com/camsiabor/qstock/dict"
	"os"
	"strings"
	"time"
)



func main() {

	var version = "0.0.1";

	defer qerr.SimpleRecover(0);
	defer signalHandle();

	var g = global.GetInstance();
	flag.StringVar(&g.LogPath,  "log", "log", "log file path");
	flag.StringVar(&g.ConfigPath,  "config", "config.json", "configuration file path");
	flag.StringVar(&g.TimeZone,  "timezone", "Asia/Shanghai", "timezone");
	flag.StringVar(&g.Mode,  "mode", dict.MODE_MASTER, "run mode");

	time.LoadLocation(g.TimeZone)

	g.Continue = true;
	g.Version = version;
	g.PanicHandler = func(pan interface{}) {
		qlog.Log(qlog.ERROR, pan);
	}
	g.SetData("global", g);
	g.SetData("u", wrap.U);
	g.SetData("cachem", scache.GetCacheManager());
	g.Run();

	var doHelp = flag.Bool( "help", false, "help");
	var doVersion = flag.Bool( "version", false, "version");
	flag.Parse();
	if (*doVersion) {
		fmt.Println("version");
	}
	if (*doHelp) {
		flag.PrintDefaults();
		os.Exit(0);
	}

	g.Mode = strings.ToLower(g.Mode);
	var logi = qlog.GetLogManager().GetDef();
	switch g.Mode {
	case dict.MODE_MASTER:
		logi.FilePrefix = "m";
	case dict.MODE_DAEMON:
		logi.FilePrefix = "d";
	}

	var workingDir, _ = os.Getwd();
	qlog.Log(qlog.INFO, g.Mode, "init", workingDir);

	var pidfilename = g.Mode + ".pid";
	var pidfilelock = qos.NewFileLock(pidfilename);
	if err := pidfilelock.Lock(); err != nil {
		qlog.Log(qlog.ERROR, g.Mode, "pid file locked", pidfilename);
		os.Exit(1);
	}
	pidfilelock.WriteString(os.Getpid());
	defer pidfilelock.UnLock();


	// [Config] ------------------------------------------------------------------------------------------------
	if (len(g.ConfigPath) == 0) {
		g.ConfigPath = "config.json"
	}
	var config, err = qconfig.ConfigLoad(g.ConfigPath, "includes")
	if err != nil {
		qlog.Log(qlog.FATAL, "config", "load failure", g.ConfigPath, err)
		return
	}
	g.Config = config;

	switch g.Mode {
	case dict.MODE_MASTER:
		master(g);
	case dict.MODE_DAEMON:
		daemon(g);
	}
}

func signalHandle() {

}