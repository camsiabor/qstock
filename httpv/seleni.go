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
	o.Port = util.GetInt(opts, 9999, "port")
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

func (o *Seleni) Init() (service *selenium.Service, driver selenium.WebDriver, err error) {

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
	if err == nil {
		caps := selenium.Capabilities{"browserName": o.Type}
		if o.Type == "chrome" {
			caps.AddChrome(chrome.Capabilities{
				Args: []string{"headless"},
			})
		} else {
			// TODO
		}
		driver, err = selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", o.Port))
		if err != nil {
			defer service.Stop()
		}
	}
	return service, driver, err
}

func (o *Seleni) Get(opts []map[string]interface{}, nicemilli int) ([]map[string]interface{}, error) {
	var service, driver, err = o.Init()
	if err != nil {
		return opts, err
	}
	defer service.Stop()
	defer driver.Quit()
	if err != nil {
		return opts, err
	}
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
