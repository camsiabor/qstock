-- http://data.10jqka.com.cn/funds/ggzjl/
-- http://data.10jqka.com.cn/funds/ggzjl/field/zjjlr/order/desc/page/1/ajax/1/


local global = require("q.global")
local simple = require("common.simple")
local stock_group = require("sync.th.mod_stcok_group")
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
opts.request_to   = 2

opts.persist = true

opts.concurrent = simple.get(profile, "concurrent", 3)

if global.runtime.GOOS() == "windows" then
    opts.newsession = false
else
    opts.newsession = true
end

opts.db = "group"
opts.datasrc = "th"

opts.print_data_from = -1
opts.print_data_to = -1

stock_group_inst:go(opts)