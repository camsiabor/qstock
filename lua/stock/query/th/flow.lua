-- http://data.10jqka.com.cn/funds/ggzjl/
-- http://data.10jqka.com.cn/funds/ggzjl/field/zjjlr/order/desc/page/1/ajax/1/


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
opts.date_offset_from = -2
opts.date_offset_to = 0

opts.db = "flow"
opts.datasrc = "th"
opts.field = "zjjlr"
opts.order = "desc"

opts.link_stock_group = true

--opts.sort_field = "flow_io_rate"
opts.sort_field = "flow_big_rate_cross_ex"


local names_bought = {
    "大名城", "生益科技", "中南传媒", "成都银行", "天龙股份", "美好置业", "科士达", "棒杰股份", "凯文教育", "三夫户外", "科力尔"
}

local names_sold = {
    "安信信托", "吉视传媒", "全筑股份", "中广天择", "黑芝麻", "瑞康医药", "天山股份"
}

local tobe_sold = {
    "鲁信创投", "西部证券", "陕国投A"
}

local names_specific = {
    "中铝国际", "TCL集团"
}
opts.filters =  {
    
    -- codes
    --filters.codes({  codes = codes_bought })
    
    --names
    --filters.names({  names = names_sold })
    --filters.names({  names = names_specific })
    
    --groups
    filters.groups( { groups = { "一带一路" } } ),
    
    --moderate
    --filters.io({  io_lower = 1.35, io_upper = 1.75, ch_lower = 1, ch_upper = 3.5, big_in_lower = 35  }),
    
    

    --high io
    --filters.io({  io_lower = 2, io_upper = 100, ch_lower = -1.5, ch_upper = 6.5, big_in_lower = 10  })
    
    -- flow in increase
    --filters.io_increase({ in_lower = 55, in_upper = 100, in_swing = 3, ch_lower = -3, ch_upper = 7 })
    

}

th_mod_flow_inst:go(opts)