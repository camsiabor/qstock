-- http://data.10jqka.com.cn/funds/ggzjl/
-- http://data.10jqka.com.cn/funds/ggzjl/field/zjjlr/order/desc/page/1/ajax/1/
--http://news.10jqka.com.cn/20190322/c610418561.shtml#refCountId=pop_50ab41b8_259

local th_mod_flow = require("sync.th.mod_flow")
local filters = require("sync.th.mod_flow_filters")

local th_mod_flow_inst = th_mod_flow:new()

local opts = {}

opts.debug = false
opts.loglevel = 0
opts.browser = "firefox"

opts.request = false
opts.request_from = 1
opts.request_to = 71
opts.nice = 0

opts.concurrent = 3
opts.newsession = false
opts.persist = true

opts.date_offset = 0
opts.date_offset_from = -6
opts.date_offset_to = 10

opts.db = "flow"
opts.datasrc = "th"
opts.field = "zjjlr"
opts.order = "desc"

opts.link_stock_group = true

--opts.sort_field = "flow_io_rate"
opts.sort_field = "flow_big_rate_cross_ex"


local names_in_list = {
    "中设股份", -- io > 2
    "沃特股份", -- io > 1.7
    "凯文教育",
    "长城影视", -- big swing
    "全筑股份", -- big force
    "中广天择",
    "普利特", -- > 2
    "中国一重",
    "延安必康"
}

local names_bought = {
    "大名城", "生益科技", "中南传媒", "成都银行", "宁波银行", 
    "科士达", "凯文教育", "三夫户外", "奥瑞德", "双象股份",
    "华东重机", "柘中股份", "中航沈飞", "中国一重", "中设股份",
    "海亮股份", "中核钛白"
}

local names_sold = {
    "天龙股份", "美好置业", "棒杰股份", "科力尔", 
}

local names_sold_2 = {
    "安信信托", "吉视传媒", "全筑股份", "中广天择", "黑芝麻", "瑞康医药", "天山股份"
}


local tobe_sold = {
    "鲁信创投", "西部证券", "陕国投A"
}

local names_specific = {
    "中铝国际", "TCL集团", "博信股份", "普利特", "中水渔业",
}

local names_maybe = {
    "杰瑞股份", -- 一带一路, io > 2
    "巨星科技", -- 堆土给
}

local codes_list = { 
    "002440", "002824", "000610", "600212", "000721",
    "002243", "000677", "002908", "603801", "002888", "002246",
}


opts.filters = {
    
    filters.no3(),
    
    
    
    --names
    --filters.names({  names = names_sold })
    --filters.names({  names = names_sold })
    --filters.names({  names = names_specific })
    --filters.names({  names = names_maybe })
    
    --codes
    filters.codes({ codes = codes_list })
    
    --groups
    --filters.groups( { groups = { "一带一路", "军工", "军民融合", "一带一路", "国产航母", "海工装备" } } ),
    
    --moderate
    --filters.io({  io_lower = 1.25, io_upper = 100, ch_lower = 0, ch_upper = 10, big_in_lower = 10  }),
    
    --high io
    --filters.io({  io_lower = 1.7, io_upper = 100, ch_lower = 0, ch_upper = 6.5, big_in_lower = 10  })
    
    --very high io
    --filters.io({  io_lower = 1.85, io_upper = 100, ch_lower = 0, ch_upper = 10, big_in_lower = 0 })
    
    -- flow in increase
    --filters.io_increase({ in_lower = 30, in_upper = 100, in_swing = 5, ch_lower = -3, ch_upper = 10 })
    
    -- chase high 
    --filters.io_any({  io_lower = 1.5, io_upper = 100, ch_lower = 5, ch_upper = 10, big_in_lower = 0 }),
    
    -- find underline
    --filters.io_any({  io_lower = 2, io_upper = 100, ch_lower = 0, ch_upper = 10, big_in_lower = 0  }),
    
}


th_mod_flow_inst:go(opts)