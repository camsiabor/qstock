package httpv

import (
	"fmt"
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

type Seleni struct {
	Name         string
	Config       map[string]interface{}
	Type         string
	DriverPath   string
	SeleniumPath string
	Port         int
	Output       io.Writer
	service      *selenium.Service
	Headless     bool
}

func (o *Seleni) initDefault() {

	var opts = o.Config
	o.Type = util.GetStr(opts, "firefox", "type")
	o.Port = util.GetInt(opts, 60001, "port")
	o.SeleniumPath = util.GetStr(opts, "selenium-server.jar", "path")
	o.DriverPath = util.GetStr(opts, "", "driver")
	o.Headless = util.GetBool(opts, true, "headless")

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

func (o *Seleni) InitService() (service *selenium.Service, err error) {

	o.initDefault()

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
	o.service, err = selenium.NewSeleniumService(o.SeleniumPath, o.Port, opts...)
	return o.service, err
}

func (o *Seleni) InitDriver() (driver selenium.WebDriver, err error) {

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
	var url = fmt.Sprintf("http://localhost:%d/wd/hub", o.Port)
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

func (o *Seleni) InitDrivers(count int) (drivers []selenium.WebDriver, err error) {
	drivers = make([]selenium.WebDriver, count)

	defer func() {
		var pan = recover()
		if pan != nil {
			o.ReleaseDrivers(drivers)
		}
	}()
	for i := 0; i < count; i++ {
		drivers[i], err = o.InitDriver()
		if err != nil {
			panic(err)
		}
	}
	return drivers, err
}

func (o *Seleni) ReleaseDrivers(drivers []selenium.WebDriver) {
	if drivers == nil {
		return
	}
	for i := 0; i < len(drivers); i++ {
		if drivers[i] != nil {
			drivers[i].Quit()
		}
	}
}

func (o *Seleni) Get(opts []map[string]interface{}, nicemilli int) ([]map[string]interface{}, error) {
	var driver, err = o.InitDriver()
	if err != nil {
		return opts, err
	}
	defer func() {
		driver.Quit()
	}()
	return o.GetBySameSession(driver, opts, nicemilli)
}

func (o *Seleni) GetEx(opts []map[string]interface{}, nicemilli int) ([]map[string]interface{}, error) {
	var driver, err = o.InitDriver()
	if err != nil {
		return opts, err
	}
	defer func() {
		driver.Quit()
	}()
	return o.GetByNewSession(driver, opts, nicemilli)
}

func (o *Seleni) GetBySameSession(driver selenium.WebDriver, opts []map[string]interface{}, nicemilli int) ([]map[string]interface{}, error) {
	var n = len(opts)
	for i := 0; i < n; i++ {
		var one = opts[i]
		var url = util.AsStr(one["url"], "")
		var errget = driver.Get(url)
		if errget != nil {
			one["err"] = errget
			continue
		}
		html, errget := driver.PageSource()
		if errget == nil {
			one["content"] = html
		} else {
			one["err"] = errget
		}
		if nicemilli > 0 {
			time.Sleep(time.Duration(nicemilli) * time.Millisecond)
		}
	}

	return opts, nil
}

func (o *Seleni) GetByNewSession(driver selenium.WebDriver, opts []map[string]interface{}, nicemilli int) ([]map[string]interface{}, error) {
	var n = len(opts)
	for i := 0; i < n; i++ {
		var one = opts[i]
		var url = util.AsStr(one["url"], "")
		var errget = driver.Get(url)
		if errget != nil {
			one["err"] = errget
			continue
		}
		html, errget := driver.PageSource()
		if errget == nil {
			one["content"] = html
		} else {
			one["err"] = errget
		}
		driver.Quit()
		if nicemilli > 0 {
			time.Sleep(time.Duration(nicemilli) * time.Millisecond)
		}
		if i == n-1 {
			break
		}
		driver.NewSession()
	}
	return opts, nil
}

func (o *Seleni) GetConcurrent(opts []map[string]interface{}, nicemilli int, forkcount int, newsession bool) ([]map[string]interface{}, error) {
	var optscount = len(opts)
	if forkcount > optscount {
		forkcount = optscount
	}
	var drivers, err = o.InitDrivers(forkcount)
	defer o.ReleaseDrivers(drivers)
	if err != nil {
		return opts, err
	}

	var driveropts = make([][]map[string]interface{}, forkcount)
	for i := 0; i < forkcount; i++ {
		driveropts[i] = make([]map[string]interface{}, 1)
		driveropts[i][0] = opts[i]
	}

	var n = 0
	for i := forkcount; i < optscount; i++ {
		var opt = opts[i]
		driveropts[n] = append(driveropts[n], opt)
		if n >= forkcount {
			n = 0
		}
	}

	var waitgroup sync.WaitGroup
	waitgroup.Add(forkcount)
	for i := 0; i < forkcount; i++ {
		var driver = drivers[i]
		var driveropt = driveropts[i]
		go func(driver selenium.WebDriver, driveropt []map[string]interface{}) {
			defer waitgroup.Done()
			if newsession {
				o.GetByNewSession(driver, driveropt, nicemilli)
			} else {
				o.GetBySameSession(driver, driveropt, nicemilli)
			}
		}(driver, driveropt)
	}

	waitgroup.Wait()
	return opts, err
}

func (o *Seleni) Terminate() error {
	if o.service != nil {
		return o.service.Stop()
	}
	return nil
}
