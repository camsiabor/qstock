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
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qcom/wrap"
	"github.com/camsiabor/qstock/dict"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func MainTest() {
	main()
}

func main() {

	var version = "0.0.1"

	defer qerr.SimpleRecover(0)

	var g = global.GetInstance()
	flag.StringVar(&g.LogPath, "log", "log", "log file path")
	flag.StringVar(&g.ConfigPath, "config", "config.json", "configuration file path")
	flag.StringVar(&g.TimeZone, "timezone", "Asia/Shanghai", "timezone")
	flag.StringVar(&g.Mode, "mode", dict.MODE_MASTER, "run mode")

	time.LoadLocation(g.TimeZone)

	g.Continue = true
	g.Version = version
	g.PanicHandler = func(pan interface{}) {
		qlog.Log(qlog.ERROR, pan)
	}
	g.SetData("global", g)
	g.SetData("u", wrap.U)
	g.SetData("runtime", qos.GetInfo())
	g.SetData("cachem", scache.GetManager())
	g.Run()

	signalHandle(g)

	var doHelp = flag.Bool("help", false, "help")
	var doVersion = flag.Bool("version", false, "version")
	flag.Parse()
	if *doVersion {
		fmt.Println("version")
	}
	if *doHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	g.Mode = strings.ToLower(g.Mode)
	var logi = qlog.GetLogManager().GetDef()
	switch g.Mode {
	case dict.MODE_MASTER:
		logi.FilePrefix = "m"
	case dict.MODE_DAEMON:
		logi.FilePrefix = "d"
	}

	var workingDir, _ = os.Getwd()
	qlog.Log(qlog.INFO, g.Mode, "init", workingDir)

	var pidfilename = g.Mode + ".pid"
	var pidfilelock = qos.NewFileLock(pidfilename)
	if err := pidfilelock.Lock(); err != nil {
		qlog.Log(qlog.ERROR, g.Mode, "pid file locked", pidfilename)
		os.Exit(1)
	}
	pidfilelock.WriteString(os.Getpid())

	// [Config] ------------------------------------------------------------------------------------------------
	if len(g.ConfigPath) == 0 {
		g.ConfigPath = "config.json"
	}
	var config, err = qconfig.ConfigLoad(g.ConfigPath, "includes")
	if err != nil {
		qlog.Log(qlog.FATAL, "config", "load failure", g.ConfigPath, err)
		return
	}
	g.Config = config

	switch g.Mode {
	case dict.MODE_MASTER:
		master(g)
	case dict.MODE_DAEMON:
		daemon(g)
	}

	go heartbeat(g)

	cmdHandle(g, pidfilelock)

}

func cmdHandle(g *global.G, pidfilelock *qos.FileLock) {

	if g.CycleHandler != nil {
		g.CycleHandler("", g, nil)
	}

	var ok bool
	var direct string
	for {
		direct, ok = g.WaitDirect()
		if !ok {
			direct = "exit"
			break
		}
		qlog.Log(qlog.INFO, "global direct: ", direct)
		switch direct {
		case "exit", "quit", "restart":
			break
		case "config_reload":
			// TODO
		}
	}
	pidfilelock.UnLock()

	var callback func()
	switch direct {
	case "restart":
		callback = func() {
			qlog.Log(qlog.INFO, "fork")
			qos.Fork()
		}
	case "exit", "quit":
		callback = func() {
			qlog.Log(qlog.INFO, "exit")
			os.Exit(0)
		}
	}
	if callback != nil {
		go func() {
			time.Sleep(time.Second * time.Duration(10))
			callback()
		}()
	}
	var err = g.Terminate()
	if err != nil {
		qlog.Log(qlog.ERROR, "global terminate error", err)
	}

	if callback != nil {
		callback()
	}

	qlog.Log(qlog.INFO, g.Mode, "fin")

}

func signalHandle(g *global.G) {
	var signalChannel = make(chan os.Signal, 16)
	signal.Notify(signalChannel,
		syscall.SIGSEGV, syscall.SIGQUIT,
		syscall.SIGTERM, syscall.SIGBUS, syscall.SIGABRT)

	go func() {

		for sig := range signalChannel {
			qlog.Log(qlog.INFO, "signal receive :", sig.String())
			switch sig {
			case syscall.SIGTERM, syscall.SIGQUIT:
				g.SendDirect("exit")
			case syscall.SIGSEGV, syscall.SIGABRT, syscall.SIGBUS:
				qlog.Log(qlog.FATAL, sig.String())
				g.SendDirect("restart")
			}

		}
	}()
}

func heartbeat(g *global.G) {

	var heartbeat_interval = util.GetInt(g.Config, 180, g.Mode, "heartbeat")
	var ticker = time.NewTicker(time.Second * time.Duration(heartbeat_interval))
	for range ticker.C {
		if !g.Continue {
			break
		}
		qlog.Log(qlog.INFO, "@")
	}

}
