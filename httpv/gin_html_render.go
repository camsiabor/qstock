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
	Name       string
	Path       string
	Content    []byte
	ModifyTime int64
}

type CustomHTMLRenderer struct {
	Name string
	Data interface{}
}

var htmlContentTypes = []string{"text/html charset=utf-8"}
var htmlCache = make(map[string]*HtmlInfo)
var cacheMutex = sync.RWMutex{}

func (r CustomHTMLRenderer) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = htmlContentTypes
	}
}

func (r CustomHTMLRenderer) Instance(name string, data interface{}) render.Render {
	return CustomHTMLRenderer{
		Name: name,
		Data: data,
	}
}

func (r CustomHTMLRenderer) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	var err error
	var content []byte

	cacheMutex.RLock()
	var info = htmlCache[r.Name]
	cacheMutex.RUnlock()

	if info == nil {
		var filename = r.Name
		var ginv = r.Data.(*HttpServer)
		var filepath = ginv.Rootabs + "/page/" + filename
		content, err = ioutil.ReadFile(filepath)
		if err != nil {
			w.Write([]byte(err.Error()))
			return nil
		}
		var stat, _ = os.Stat(filepath)
		info = new(HtmlInfo)
		info.Name = r.Name
		info.Path = filepath
		info.Content = content
		info.ModifyTime = stat.ModTime().Unix()
		cacheMutex.Lock()
		htmlCache[r.Name] = info
		cacheMutex.Unlock()
	} else {
		content = info.Content
	}
	w.Write(content)
	return nil
}

func GinRefreshPage(refreshInterval int) {
	for {
		time.Sleep(time.Duration(refreshInterval) * time.Second)
		cacheMutex.RLock()
		for _, one := range htmlCache {
			stat, err := os.Stat(one.Path)
			if err == nil {
				var newmod = stat.ModTime().Unix()
				if newmod > one.ModifyTime {
					content, err := ioutil.ReadFile(one.Path)
					if err == nil {
						one.Content = content
						one.ModifyTime = newmod
					}
				}
			}
		}
		cacheMutex.RUnlock()
	}
}
