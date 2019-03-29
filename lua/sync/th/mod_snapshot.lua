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


--[[
local inst = M:new()
local dates = inst:dates(3, 0, 0)
local ks = inst:snapshot("ch000009", dates)
print(ks)
]]--


return M