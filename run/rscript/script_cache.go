package rscript

import (
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/qdao"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/dict"
	"time"
)

func loadScriptByName(cache *scache.SCache, factor int, timeout time.Duration, lock bool, keys ...interface{}) (interface{}, error) {
	dao, err := qdao.GetManager().Get(cache.Dao)
	if err != nil {
		return nil, err
	}
	var name = keys[0]
	script, err := dao.Get(cache.Db, cache.Group, name, 1, nil)
	if err != nil {
		return nil, err
	}
	var meta = &Meta{}
	meta.FromMap(util.AsMap(script, false))
	return meta, nil
}

func updateScriptByName(cache *scache.SCache, flag int, val interface{}, lock bool, keys ...interface{}) error {

	if flag&scache.FLAG_UPDATE_ASPECT_BEFORE == 0 {
		return nil
	}

	dao, err := qdao.GetManager().Get(cache.Dao)
	if err != nil {
		return err
	}

	/*
		old, _ := cache.GetEx(false, 0, 0, keys[0])
		if old != nil {
			var oldmeta = val.(*Meta)
			oldmeta.Lines = nil
			oldmeta.Binary = nil
			if len(oldmeta.Hash) > 0 {
				cache.Twin.Delete(oldmeta.Hash)
			}
		}
	*/

	if flag&scache.FLAG_UPDATE_SET > 0 {
		var meta *Meta = val.(*Meta)
		val = meta.ToMap()
		_, err = dao.Update(cache.Db, cache.Group, keys[0], val, true, -1, nil)
	} else if flag&scache.FLAG_UPDATE_DELETE > 0 {
		_, err = dao.Delete(cache.Db, cache.Group, keys[0], nil)
	}

	return err
}

func clearScriptByHash() {
	var cache_script_by_hash = scache.GetManager().Get(dict.CACHE_SCRIPT_BY_HASH)
	var ticker = time.Tick(time.Hour * 6)
	for {
		<-ticker
		var keys, _ = cache_script_by_hash.Keys(true)
		if keys != nil {
			for _, key := range keys {
				cache_script_by_hash.Delete(key, true)
			}
		}
	}
}

func InitScript(g *global.G) {
	var cache_script_by_name = scache.GetManager().Get(dict.CACHE_SCRIPT_BY_NAME)
	var cache_script_by_hash = scache.GetManager().Get(dict.CACHE_SCRIPT_BY_HASH)
	cache_script_by_name.Twin = cache_script_by_hash
	cache_script_by_name.Dao = dict.DAO_MAIN
	cache_script_by_name.Db = dict.DB_COMMON
	cache_script_by_name.Group = dict.GROUP_SCRIPT

	cache_script_by_name.Loader = loadScriptByName
	cache_script_by_name.Updater = updateScriptByName

	go clearScriptByHash()
}
