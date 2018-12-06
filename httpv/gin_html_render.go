package httpv

import (
	"github.com/gin-gonic/gin/render"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)


type HtmlInfo struct {
	Name string;
	Path string;
	Content []byte;
	ModifyTime int64;
}

type CustomHTMLRenderer struct {
	Name     string
	Data     interface{}
}

var _html_content_type = []string{"text/html; charset=utf-8"}
var _html_cache = make(map[string]*HtmlInfo);
var _cache_mutex = sync.RWMutex{};


func (r CustomHTMLRenderer) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = _html_content_type;
	}
}

func (r CustomHTMLRenderer) Instance(name string, data interface{}) render.Render {
	return CustomHTMLRenderer{
		Name : name,
		Data : data,
	};
}

func (r CustomHTMLRenderer) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	var err error;
	var content []byte;

	_cache_mutex.RLock();
	var info = _html_cache[r.Name];
	_cache_mutex.RUnlock();

	if (info == nil) {
		var filename = r.Name;
		var ginv *HttpServer = r.Data.(*HttpServer);
		var filepath = ginv.Rootabs + "/page/" + filename;
		content, err = ioutil.ReadFile(filepath)
		if (err != nil) {
			w.Write([]byte(err.Error()))
			return nil;
		}
		var stat, _ = os.Stat(filepath);
		info = new(HtmlInfo);
		info.Name = r.Name;
		info.Path = filepath;
		info.Content = content;
		info.ModifyTime = stat.ModTime().Unix();
		_cache_mutex.Lock();
		_html_cache[r.Name] = info;
		_cache_mutex.Unlock();
	} else {
		content = info.Content;
	}
	w.Write(content);
	return nil;
}

func GinRefreshPage(refresh_interval int) {
	for {
		time.Sleep(time.Duration(refresh_interval) * time.Second)
		_cache_mutex.RLock();
		defer _cache_mutex.RUnlock();
		for _, one := range _html_cache {
			stat, err := os.Stat(one.Path);
			if (err == nil) {
				var newmod = stat.ModTime().Unix();
				if (newmod > one.ModifyTime) {
					content, err := ioutil.ReadFile(one.Path)
					if (err == nil) {
						one.Content = content;
						one.ModifyTime = newmod;
					}
				}
			}
		}
	}
}
