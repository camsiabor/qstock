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
	Name         string
	Type         string
	DriverPath   string
	RemotePath   string
	RemotePort   int
	Output       io.Writer
	Headless     bool
	service      *selenium.Service
	simpleClient *qnet.SimpleHttp
	Config       map[string]interface{}
}

func (o *HttpAgent) initDefault() {

	var opts = o.Config
	o.Type = util.GetStr(opts, "firefox", "type")
	o.RemotePort = util.GetInt(opts, 60001, "port")
	o.RemotePath = util.GetStr(opts, "selenium-server.jar", "path")
	o.DriverPath = util.GetStr(opts, "", "driver")
	o.Headless = util.GetBool(opts, true, "headless")

	if o.Type == "wget" {
		o.simpleClient = qnet.GetSimpleHttp()
		return
	}

	if o.DriverPath == "" {
		if o.Type == "chrome" {
			o.DriverPath = "chromedriver"
		} else {
			o.DriverPath = "geckodriver"
		}
	}

	if runtime.GOOS == "windows" {
		if !strings.Contains(o.DriverPath, ".exe") {
			o.DriverPath = o.DriverPath + ".exe"
		}
	}
	// TODO other os

}

func (o *HttpAgent) InitService() (service *selenium.Service, err error) {

	o.initDefault()

	if o.Type == "wget" {
		return
	}

	defer func() {
		var pan = recover()
		if pan != nil {
			if service != nil {
				service.Stop()
			}
			panic(pan)
		}
	}()

	if o.Output == nil {
		o.Output = os.Stdout
	}
	var browserOption selenium.ServiceOption
	if o.Type == "chrome" {
		browserOption = selenium.ChromeDriver(o.DriverPath)
	} else {
		browserOption = selenium.GeckoDriver(o.DriverPath)
	}
	opts := []selenium.ServiceOption{
		//selenium.StartFrameBuffer(), // Start an X frame buffer for the browser to run in.
		browserOption,
		selenium.Output(o.Output), // Output debug information to STDERR.
	}
	selenium.SetDebug(false)
	o.service, err = selenium.NewSeleniumService(o.RemotePath, o.RemotePort, opts...)
	return o.service, err
}

func (o *HttpAgent) InitDriver() (driver selenium.WebDriver, err error) {

	if o.Type == "wget" {
		return nil, nil
	}

	var browserName string
	if o.Type == "chrome" {
		browserName = o.Type
	} else {
		browserName = "firefox"
	}
	caps := selenium.Capabilities{"browserName": browserName}
	if o.Headless {
		if o.Type == "chrome" {
			caps.AddChrome(chrome.Capabilities{
				Args: []string{"headless"},
			})
		} else {
			caps.AddFirefox(firefox.Capabilities{
				Args: []string{"-headless"},
			})
		}
	}
	var url = fmt.Sprintf("http://localhost:%d/wd/hub", o.RemotePort)
	driver, err = selenium.NewRemote(caps, url)
	if err != nil {
		if o.service == nil {
			_, err = o.InitService()
			if err == nil {
				return o.InitDriver()
			}
		}
	}
	return driver, err
}

func (o *HttpAgent) InitDrivers(count int) (drivers []selenium.WebDriver, err error) {
	drivers = make([]selenium.WebDriver, count)

	defer func() {
		if err != nil {
			o.ReleaseDrivers(drivers)
		}
	}()

	var waitgroup = sync.WaitGroup{}
	for i := 0; i < count; i++ {
		go func() {
			defer waitgroup.Done()
			var errone error
			drivers[i], errone = o.InitDriver()
			if errone != nil && err == nil {
				err = errone
			}
		}()
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

func (o *HttpAgent) Get(opts []map[string]interface{}, nicemilli int, newsession bool, concurrent int, loglevel int) ([]map[string]interface{}, error) {

	if concurrent > 1 {
		return o.GetConcurrent(opts, nicemilli, newsession, concurrent, loglevel)
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
	if newsession {
		return o.GetByNewSession(driver, opts, nicemilli, loglevel)
	} else {
		return o.GetBySameSession(driver, opts, nicemilli, loglevel)
	}
}

func (o *HttpAgent) GetBySameSession(driver selenium.WebDriver, opts []map[string]interface{}, nicemilli int, loglevel int) ([]map[string]interface{}, error) {
	var n = len(opts)

	for i := 0; i < n; i++ {
		var html string
		var errget error
		var one = opts[i]
		var url = util.AsStr(one["url"], "")
		if o.Type == "wget" {
			var encoding = util.GetStr(one, "utf-8", "encoding")
			var headers = util.GetStringMap(one, false, "headers")
			html, _, errget = o.simpleClient.Get(url, headers, encoding)
		} else {
			errget = driver.Get(url)
			if errget == nil {
				html, errget = driver.PageSource()
			}
		}
		if errget == nil {
			one["content"] = html
			if loglevel >= 0 {
				qlog.Log(qlog.INFO, "httpagent", "success", url)
			}
		} else {
			one["err"] = errget
			if loglevel >= 0 {
				qlog.Log(qlog.INFO, "httpagent", "fail", url)
			}
		}

		if nicemilli > 0 {
			time.Sleep(time.Duration(nicemilli) * time.Millisecond)
		}

	}

	return opts, nil
}

func (o *HttpAgent) GetByNewSession(driver selenium.WebDriver, opts []map[string]interface{}, nicemilli int, loglevel int) ([]map[string]interface{}, error) {
	var n = len(opts)
	for i := 0; i < n; i++ {
		var html string
		var errget error
		var one = opts[i]
		var url = util.AsStr(one["url"], "")
		if o.Type == "wget" {
			var encoding = util.GetStr(opts, "utf-8", "encoding")
			var headers = util.GetStringMap(opts, false, "headers")
			html, _, errget = o.simpleClient.Get(url, headers, encoding)
		} else {
			errget = driver.Get(url)
			if errget == nil {
				html, errget = driver.PageSource()
			}
		}
		if errget == nil {
			one["content"] = html
			if loglevel >= 0 {
				qlog.Log(qlog.INFO, "httpagent", "success", url)
			}
		} else {
			one["err"] = errget
			if loglevel >= 0 {
				qlog.Log(qlog.INFO, "httpagent", "fail", url, errget.Error())
			}
		}
		if driver != nil {
			driver.Quit()
		}
		if nicemilli > 0 {
			time.Sleep(time.Duration(nicemilli) * time.Millisecond)
		}
		if i == n-1 {
			break
		}
		if driver != nil {
			driver.NewSession()
		}
	}
	return opts, nil
}

func (o *HttpAgent) GetConcurrent(opts []map[string]interface{}, nicemilli int, newsession bool, concurrent int, loglevel int) ([]map[string]interface{}, error) {
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
						qlog.Log(qlog.INFO, "httpagent", "one concurrent done", index)
					}
				} else {
					if loglevel >= 0 {
						qlog.Log(qlog.ERROR, "httpagent", "one concurrent error", pan)
					}
				}
			}()
			if newsession {
				o.GetByNewSession(driver, driveropt, nicemilli, loglevel)
			} else {
				o.GetBySameSession(driver, driveropt, nicemilli, loglevel)
			}
		}(driver, driveropt, i)
	}
	waitgroup.Wait()
	if loglevel >= 0 {
		qlog.Log(qlog.INFO, "httpagent", "concurrent fin", concurrent)
	}
	return opts, err
}

func (o *HttpAgent) Terminate() error {
	if o.service != nil {
		return o.service.Stop()
	}
	return nil
}
