-- http://data.10jqka.com.cn/funds/ggzjl/
-- http://data.10jqka.com.cn/funds/ggzjl/field/zjjlr/order/desc/page/1/ajax/1/

--os.exit(1)

local global = require("q.global")
local loggerm = require("q.logger")
local simple = require("common.simple")
local stock_group = require("sync.th.mod_stock_group")
local stock_group_inst = stock_group:new()


local work = global.work
local profile
if work == nil then
    profile = { }
else
    profile = work.Profile
end

local opts = { }

opts.debug = false
opts.loglevel = 0
opts.browser = "firefox"
--opts.browser = "gorilla"

opts.request = true
opts.request_from = 1
opts.request_to   = 10
opts.request_types = { "concept", "industry" }
opts.request_types = { "industry" }
--opts.request_types = { "concept" }

opts.reload_check = true

opts.persist = true

opts.concurrent = simple.get(profile, "concurrent", 1)

if global.runtime.GOOS() == "windows" then  
    opts.newsession = false
else
    opts.newsession = true
end

opts.db = "group"
opts.datasrc = "th"

opts.print_data_from = -1
opts.print_data_to = -1


--local url = stock_group_inst:get_url_pattern("concept", "group")
--print(url)
local groups = stock_group_inst:goes(opts)
if groups == nil then
    print("[stock.group] null")
else 
    print("[stock.group] reload count", simple.table_count(groups))
end

--[[
for code, group in pairs(groups) do 
    print(code, group.name, "   ", group.page, simple.table_count(group.list))
end
]]--
--simple.table_print_all(groups)