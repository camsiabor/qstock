local M = {}
M.__index = M

local global = require("q.global")
local json = require('common.json')
local simple = require("common.simple")

-------------------------------------------------------------------------------------------

function M:new()
    local inst = {}
    inst.__index = self
    setmetatable(inst, self)
    return inst
end

-------------------------------------------------------------------------------------------

function M:dates(from, date_offset, to, doreverse) 
    if doreverse == nil then
        doreverse = false
    end
    local dates = global.calendar.List(from, date_offset, to, doreverse)
    return dates
end

function M:snapshot(code, dates)
    if self.cache_stock_khistory == nil then
        self.cache_stock_khistory = global.cachem.Get("stock.khistory")
    end
    local cache = self.cache_stock_khistory
    local ks = cache.ListSubVal(true, { code }, dates)
    return ks
end

function M:merge(serie, k)

    if serie == nil or k == nil then
        return
    end

    serie.date = k.date
    serie.open = k.open + 0
    serie.close = k.close
    if serie.close == nil then
        serie.close = k.open
    end
    serie.close = serie.close + 0

    serie.pre_close = k.pre_close + 0

    serie.swing = k.swing + 0
    serie.low = k.low + 0
    serie.high = k.high + 0

    serie.pb = k.pb + 0

    return serie
end

function M:merges(series, ks)
    if series == nil or ks == nil then
        return nil
    end
    local n = #series
    for s = 1, n do
        local serie = series[s]
        local k = ks[s]
        self:merge(serie, k)
    end
    return series
end

--[[
local inst = M:new()
local dates = inst:dates(3, 0, 0)
local ks = inst:snapshot("ch000009", dates)
print(ks)
]]--


return M