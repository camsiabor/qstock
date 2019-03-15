package httpv

import (
	"fmt"
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/util"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

type Seleni struct {
	Type         string
	DriverPath   string
	SeleniumPath string
	Port         int
	Output       io.Writer
}

func (o *Seleni) initDefault() {
	var g = global.GetInstance()
	var config = g.Config
	var opts = util.GetMap(config, true, "selenium")
	o.Type = util.GetStr(opts, "chrome", "type")
	o.Port = util.GetInt(opts, 60000, "port")
	o.SeleniumPath = util.GetStr(opts, "selenium-server.jar", "path")
	o.DriverPath = util.GetStr(opts, "", "driver")

	// TODO port check

	if o.DriverPath == "" {
		if o.Type == "chrome" {
			if runtime.GOOS == "windows" {
				o.DriverPath = "chromedriver.exe"
			} else {
				o.DriverPath = "chromedriver"
			}
		} else {
			// TODO
		}
	}
}

func (o *Seleni) Init(dcount int) (service *selenium.Service, drivers []selenium.WebDriver, err error) {

	o.initDefault()

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

	service, err = selenium.NewSeleniumService(o.SeleniumPath, o.Port, opts...)
	if err != nil {
		if service != nil {
			service.Stop()
		}
		return nil, nil, err
	}

	caps := selenium.Capabilities{"browserName": o.Type}
	if o.Type == "chrome" {
		caps.AddChrome(chrome.Capabilities{
			Args: []string{"headless"},
		})
	} else {
		// TODO
	}

	var haserr bool = false
	drivers = make([]selenium.WebDriver, dcount)
	for i := 0; i < dcount; i++ {
		drivers[i], err = selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", o.Port+i))
		if err != nil {
			haserr = true
			break
		}
	}

	if haserr {
		for i := 0; i < dcount; i++ {
			var driver = drivers[i]
			if driver != nil {
				driver.Quit()
				driver.Close()
			}
		}
		service.Stop()
		service = nil
		drivers = nil
	}

	return service, drivers, err
}

func (o *Seleni) GetPrimary(opts []map[string]interface{}, nicemilli int) ([]map[string]interface{}, error) {
	var service, drivers, err = o.Init(1)
	if err != nil {
		return opts, err
	}
	defer service.Stop()

	var driver = drivers[0]
	defer driver.Close()
	defer driver.Quit()

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
		if errget != nil {
			one["err"] = errget
			continue
		}
		one["content"] = html
		time.Sleep(time.Duration(nicemilli) * time.Millisecond)
	}
	return opts, err
}

func (o *Seleni) Get(opts []map[string]interface{}, nicemilli int) ([]map[string]interface{}, error) {
	var service, drivers, err = o.Init(1)
	if err != nil {
		return opts, err
	}
	defer service.Stop()
	var driver = drivers[0]
	defer driver.Close()
	defer driver.Quit()
	return o.DGets(driver, opts, nicemilli)
}

func (o *Seleni) DGets(driver selenium.WebDriver, opts []map[string]interface{}, nicemilli int) ([]map[string]interface{}, error) {

	var n = len(opts)

	for i := 0; i < n; i++ {
		var one = opts[i]
		var url = util.AsStr(one["url"], "")
		func() {
			var sid, serr = driver.NewSession()
			if serr == nil {
				defer driver.Quit()
			}
			driver.SwitchSession(sid)
			var errget = driver.Get(url)
			if errget != nil {
				one["err"] = errget
				return
			}
			html, errget := driver.PageSource()
			if errget == nil {
				one["content"] = html
			} else {
				one["err"] = errget
			}
		}()
		time.Sleep(time.Duration(nicemilli) * time.Millisecond)
	}

	return opts, nil
}

func (o *Seleni) GetEx(opts []map[string]interface{}, nicemilli int, forkcount int) ([]map[string]interface{}, error) {
	var optscount = len(opts)
	if forkcount > optscount {
		forkcount = optscount
	}
	var service, drivers, err = o.Init(forkcount)
	if err != nil {
		return opts, err
	}
	defer service.Stop()
	defer func() {
		for i := 0; i < len(drivers); i++ {
			var driver = drivers[i]
			if driver != nil {
				driver.Quit()
				driver.Close()
			}
		}
	}()

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
		go func() {
			defer waitgroup.Done()
			o.DGets(driver, driveropt, nicemilli)
		}()
	}

	waitgroup.Wait()
	return opts, err
}
