package httpv

import (
	"fmt"
	"github.com/camsiabor/qcom/qlog"
	"github.com/camsiabor/qcom/qnet"
	"github.com/camsiabor/qcom/util"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"github.com/tebeka/selenium/firefox"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type HttpAgent struct {
	Name string

	atype      string
	driverPath string
	remotePath string

	remotePortIndex int
	remotePortLower int
	remotePortUpper int

	Output       io.Writer
	headless     bool
	services     []*selenium.Service
	simpleClient *qnet.SimpleHttp
	config       map[string]interface{}

	basicHttp        bool
	basicHttpChecked bool

	mutex     sync.RWMutex
	portmutex sync.RWMutex
}

func (o *HttpAgent) InitParameters(config map[string]interface{}) {

	var opts = config
	o.basicHttpChecked = false
	o.atype = util.GetStr(opts, "firefox", "type")

	o.remotePortLower = util.GetInt(opts, 60000, "port")
	o.remotePortUpper = util.GetInt(opts, 60010, "port")

	o.remotePath = util.GetStr(opts, "selenium-server.jar", "path")
	o.driverPath = util.GetStr(opts, "", "driver")
	o.headless = util.GetBool(opts, true, "headless")

	if o.IsBasicHttp() {
		o.simpleClient = qnet.GetSimpleHttp()
		return
	}

	if o.driverPath == "" {
		if o.atype == "chrome" {
			o.driverPath = "chromedriver"
		} else {
			o.driverPath = "geckodriver"
		}
	}

	if runtime.GOOS == "windows" {
		if !strings.Contains(o.driverPath, ".exe") {
			o.driverPath = o.driverPath + ".exe"
		}
	}
	// TODO other os

}

func (o *HttpAgent) IsBasicHttp() bool {
	if !o.basicHttpChecked {
		o.basicHttp = o.atype == "" || o.atype == "std" || o.atype == "gorilla"
		o.basicHttpChecked = true
	}
	return o.basicHttp
}

func (o *HttpAgent) getAvailablePort() int {
	o.portmutex.Lock()
	defer o.portmutex.Unlock()
	o.remotePortIndex = o.remotePortIndex + 1
	if o.remotePortIndex < o.remotePortLower || o.remotePortIndex > o.remotePortUpper {
		o.remotePortIndex = o.remotePortLower
	}
	return o.remotePortIndex
}

func (o *HttpAgent) InitService() ([]*selenium.Service, error) {

	o.InitParameters(o.config)

	if o.IsBasicHttp() {
		return o.services, nil
	}

	if o.services != nil {
		o.StopService()
	}

	defer func() {
		var pan = recover()
		if pan != nil {
			o.StopService()
			panic(pan)
		}
	}()

	if o.Output == nil {
		o.Output = os.Stdout
	}
	var browserOption selenium.ServiceOption
	if o.atype == "chrome" {
		browserOption = selenium.ChromeDriver(o.driverPath)
	} else {
		browserOption = selenium.GeckoDriver(o.driverPath)
	}
	opts := []selenium.ServiceOption{
		//selenium.StartFrameBuffer(), // Start an X frame buffer for the browser to run in.
		browserOption,
		selenium.Output(o.Output), // Output debug information to STDERR.
	}
	var err error
	selenium.SetDebug(false)

	if o.remotePortLower < 1000 {
		o.remotePortLower = 60000
	}
	if o.remotePortUpper <= o.remotePortLower {
		o.remotePortUpper = o.remotePortLower + 10
	}
	var count = o.remotePortUpper - o.remotePortLower
	o.services = make([]*selenium.Service, count)
	for i := 0; i < count; i++ {
		o.services[i], err = selenium.NewSeleniumService(o.remotePath, o.remotePortLower+i, opts...)
	}
	return o.services, err
}

func (o *HttpAgent) StopService() {
	if o.services == nil {
		return
	}
	o.mutex.Lock()
	defer o.mutex.Unlock()
	var n = len(o.services)
	for i := 0; i < n; i++ {
		var service = o.services[i]
		if service != nil {
			func() {
				defer recover()
				service.Stop()
			}()
		}
	}
}

func (o *HttpAgent) InitDriver() (driver selenium.WebDriver, err error) {

	if o.IsBasicHttp() {
		return nil, nil
	}

	var browserName string
	if o.atype == "chrome" {
		browserName = o.atype
	} else {
		browserName = "firefox"
	}
	caps := selenium.Capabilities{"browserName": browserName}
	if o.headless {
		if o.atype == "chrome" {
			caps.AddChrome(chrome.Capabilities{
				Args: []string{"headless"},
			})
		} else {
			caps.AddFirefox(firefox.Capabilities{
				Args: []string{"-headless"},
			})
		}
	}
	var port = o.getAvailablePort()
	var url = fmt.Sprintf("http://localhost:%d/wd/hub", port)
	driver, err = selenium.NewRemote(caps, url)
	if err != nil {
		if o.services == nil {
			o.mutex.Lock()
			defer o.mutex.Unlock()
			if o.services == nil {
				_, err = o.InitService()
			} else {
				err = nil
			}
			if err == nil {
				return o.InitDriver()
			}
		}
	}
	return driver, err
}

func (o *HttpAgent) InitDrivers(count int) (drivers []selenium.WebDriver, err error) {
	drivers = make([]selenium.WebDriver, count)

	if o.IsBasicHttp() {
		return drivers, nil
	}

	defer func() {
		if err != nil {
			o.ReleaseDrivers(drivers)
		}
	}()

	var waitgroup = sync.WaitGroup{}
	waitgroup.Add(count)
	for i := 0; i < count; i++ {
		go func(i int) {
			defer waitgroup.Done()
			var errone error
			drivers[i], errone = o.InitDriver()
			if errone != nil && err == nil {
				err = errone
			}
		}(i)
	}
	waitgroup.Wait()

	return drivers, err
}

func (o *HttpAgent) ReleaseDrivers(drivers []selenium.WebDriver) {
	if drivers == nil {
		return
	}
	for i := 0; i < len(drivers); i++ {
		var driver = drivers[i]
		if driver != nil {
			driver.Quit()
		}
	}
}

func (o *HttpAgent) Get(opts []map[string]interface{}, nicemilli int, newsession int, concurrent int, loglevel int) ([]map[string]interface{}, error) {
	if len(opts) == 0 {
		return opts, nil
	}
	if concurrent > 1 {
		if o.IsBasicHttp() {
			return o.GetSimpleConcurrent(opts, nicemilli, newsession, concurrent, loglevel)
		} else {
			return o.GetDriverConcurrent(opts, nicemilli, newsession, concurrent, loglevel)
		}

	}
	var driver, err = o.InitDriver()
	if err != nil {
		return opts, err
	}
	defer func() {
		if driver != nil {
			driver.Quit()
		}
	}()
	return o.GetMany(driver, opts, nicemilli, newsession, loglevel)
}

func (o *HttpAgent) GetOne(driver selenium.WebDriver, opt map[string]interface{}, newsession bool, loglevel int) (map[string]interface{}, error) {
	var html string
	var errget error
	var url = util.AsStr(opt["url"], "")
	if o.IsBasicHttp() {
		var encoding = util.GetStr(opt, "utf-8", "encoding")
		var headers = util.GetStringMap(opt, false, "headers")
		if o.simpleClient == nil {
			o.simpleClient = qnet.GetSimpleHttp()
		}
		html, _, errget = o.simpleClient.Get(o.atype, url, headers, encoding)
	} else {
		errget = driver.Get(url)
		if errget == nil {
			html, errget = driver.PageSource()
		}
		if newsession {
			driver.Quit()
			driver.NewSession()
		}
	}
	if errget == nil {
		opt["content"] = html
		if loglevel >= 0 {
			qlog.Log(qlog.INFO, "httpagent", o.atype, "done", url, len(html))
		}
	} else {
		opt["err"] = errget
		if loglevel >= 0 {
			qlog.Log(qlog.INFO, "httpagent", o.atype, "fail", url, errget.Error())
		}
	}
	return opt, nil
}

func (o *HttpAgent) GetMany(driver selenium.WebDriver, opts []map[string]interface{}, nicemilli int, newsession int, loglevel int) ([]map[string]interface{}, error) {
	var n = len(opts)
	var sessioncount = 0
	for i := 0; i < n; i++ {
		var one = opts[i]
		var refreshsession = false
		sessioncount = sessioncount + 1
		if newsession > 0 && sessioncount >= newsession {
			sessioncount = 0
			refreshsession = true
		}
		o.GetOne(driver, one, refreshsession, loglevel)
		if nicemilli > 0 {
			time.Sleep(time.Duration(nicemilli) * time.Millisecond)
		}

	}
	return opts, nil
}

func (o *HttpAgent) GetDriverConcurrent(opts []map[string]interface{}, nicemilli int, newsession int, concurrent int, loglevel int) ([]map[string]interface{}, error) {
	var optscount = len(opts)
	if concurrent > optscount {
		concurrent = optscount
	}
	var drivers, err = o.InitDrivers(concurrent)
	defer o.ReleaseDrivers(drivers)
	if err != nil {
		return opts, err
	}

	var driveropts = make([][]map[string]interface{}, concurrent)
	for i := 0; i < concurrent; i++ {
		driveropts[i] = make([]map[string]interface{}, 1)
		driveropts[i][0] = opts[i]
	}

	var n = 0
	for i := concurrent; i < optscount; i++ {
		var opt = opts[i]
		driveropts[n] = append(driveropts[n], opt)
		if n >= concurrent {
			n = 0
		}
	}

	var waitgroup sync.WaitGroup
	waitgroup.Add(concurrent)
	for i := 0; i < concurrent; i++ {
		var driver = drivers[i]
		var driveropt = driveropts[i]
		go func(driver selenium.WebDriver, driveropt []map[string]interface{}, index int) {
			defer func() {
				defer waitgroup.Done()
				var pan = recover()
				if pan == nil {
					if loglevel >= 0 {
						qlog.Log(qlog.INFO, "httpagent", o.atype, "one concurrent done", index)
					}
				} else {
					if loglevel >= 0 {
						var panerr, ok = pan.(error)
						if ok {
							pan = panerr.Error()
						}
						qlog.Log(qlog.ERROR, "httpagent", o.atype, "one concurrent error", pan)
					}
				}
			}()

			o.GetMany(driver, driveropt, nicemilli, newsession, loglevel)

		}(driver, driveropt, i)
	}
	waitgroup.Wait()
	if loglevel >= 0 {
		qlog.Log(qlog.INFO, "httpagent", o.atype, "driver concurrent fin", concurrent)
	}
	return opts, err
}

func (o *HttpAgent) GetSimpleConcurrent(opts []map[string]interface{}, nicemilli int, newsession int, concurrent int, loglevel int) ([]map[string]interface{}, error) {
	var optscount = len(opts)

	if concurrent > optscount {
		concurrent = optscount
	}

	var erroverall error
	var waitgroup sync.WaitGroup

	var n = 0
	for n < optscount {
		waitgroup.Add(concurrent)
		for i := 0; i < concurrent; i++ {
			if n >= optscount {
				break
			}
			var opt = opts[n]
			n = n + 1

			go func(opt map[string]interface{}) {
				defer func() {
					defer waitgroup.Done()
					var pan = recover()
					if pan != nil {
						if loglevel >= 0 {
							var panerr, ok = pan.(error)
							if ok {
								if erroverall != nil {
									erroverall = panerr
								}
								pan = panerr.Error()
							}
							qlog.Log(qlog.ERROR, "httpagent", o.atype, "one concurrent error", pan)
						}
					}
				}()
				o.GetOne(nil, opt, false, loglevel)
			}(opt)
		}
		waitgroup.Wait()
	}
	if loglevel >= 0 {
		qlog.Log(qlog.INFO, "httpagent", o.atype, "simple concurrent fin", concurrent)
	}
	return opts, erroverall
}

func (o *HttpAgent) Terminate() error {
	o.StopService()
	return nil
}
