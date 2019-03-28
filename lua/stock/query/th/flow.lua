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
opts.date_offset_from = -8
opts.date_offset_to = 10

opts.db = "flow"
opts.datasrc = "th"
opts.field = "zjjlr"
opts.order = "desc"

opts.link_stock_group = true

--opts.sort_field = "flow_io_rate"
opts.sort_field = "flow_big_rate_cross_ex"




local names_bought = {
    "贵人鸟",
    "黑芝麻",
    "美盛文化", 
    "中南传媒", 
    "华东重机", "中国一重",
    "中核钛白"
}


opts.result_adapter2 = function(opts, result, mapping)
    local up = 0
    local down = 0
    local n = #result
    for i = 1, n do 
        local one = result[i]
        local code = one.code
        local series = mapping[code]
        local near = series[2]
        
        if near ~= nil then
            if near.change_rate > 0 then
                up = up + 1
            else
                down = down + 1
            end
        end
        
    end
    print("[up]", up)
    print("[down]", down)
    print("[up/down]", (up) / (up + down) * 100)
end


opts.filters = {
    
    filters.no3(),
    
    
    --names
    --filters.names({  names = names_bought }),
    --filters.names({  names = names_sold }),
    --filters.names({  names = names_sold }),
    --filters.names({  names = names_specific }),
    --filters.names({  names = names_maybe }),
    
    --codes
    --filters.codes({ codes = codes_list })
    
    --groups
    --filters.groups( { groups = { "一带一路", "军工", "军民融合", "一带一路", "国产航母", "海工装备" } } ),
    
    --moderate
    --filters.io({  io_lower = 1.2, io_upper = 100, ch_lower = 5, ch_upper = 11, big_in_lower = 0  }),
    
    -- low io
    --filters.io({  io_lower = 0.8, io_upper = 1.25, ch_lower = 0, ch_upper = 5, big_in_lower = 0  })
    
    --high io
    --filters.io({  io_lower = 1.25, io_upper = 100, ch_lower = 0.5, ch_upper = 3, big_in_lower = 20  })
    
    --anit io
    --filters.io({  io_lower = 1.5, io_upper = 100, ch_lower = -5, ch_upper = 0, big_in_lower = 20  })
    
    --very high io
    filters.io({  io_lower = 1.75, io_upper = 100, ch_lower = -1, ch_upper = 10.5, big_in_lower = 0 }),
    
    -- io ceil
    --filters.io({  io_lower = 0, io_upper = 100, ch_lower = 9, ch_upper = 11, big_in_lower = 0 }),
    
    -- flow in increase
    --filters.io_increase({ in_lower = 30, in_upper = 100, in_swing = 5, ch_lower = -10, ch_upper = 10 })
    
    -- chase high 
    --filters.io_any({  io_lower = 1.5, io_upper = 100, ch_lower = -1, ch_upper = 5, big_in_lower = 0 }),
    
    -- find underline
    --filters.io_any({  io_lower = 1.85, io_upper = 100, ch_lower = 0, ch_upper = 10, big_in_lower = 0  }),
    
    --filters.io_any({  io_lower = 0.5, io_upper = 0.75, ch_lower = -1, ch_upper = 5, big_in_lower = 0  }),
    --filters.io_any({  io_lower = 1.75, io_upper = 100, ch_lower = -9, ch_upper = 10.5, big_in_lower = 0  }),
    
}


th_mod_flow_inst:go(opts)