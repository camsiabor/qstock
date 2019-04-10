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

opts.request_from = 1
opts.request_to = 71
opts.request_each = 1
opts.concurrent = 1
opts.newsession = 5
opts.nice = 0

opts.persist = true

opts.date_show = 10


opts.db = "flow"
opts.datasrc = "th"
opts.field = "zjjlr"
opts.order = "desc"

opts.link_stock_group = true
opts.link_stock_snapshot = false

opts.sort_field = "flow_io_rate"
--opts.sort_field = "flow_big_rate_cross_ex"


local names_target = {
    "福田汽车",
}

local names_bought = {
    "中远海能", "中集集团", "中联重科", "中源家居",
    "宁波海运", "宁波港",
    "皖江物流", "天顺股份",
    "中国武夷",
    "石化机械",
    "中国汽研", "长安汽车",
    "柳药股份", "汉森制药",
    "诺力股份",
    "长江证券",
    "山煤国际",
    "科华恒盛",
    "英联股份",
    
}

local names_tobe = {
    "嘉应制药"
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

opts.request = false

opts.date_offset = 0
opts.date_offset_to = 12
--opts.date_offset_from = -2
opts.date_offset_from = -opts.date_show - opts.date_offset


opts.filters = {
    
    filters.no3(),
    
    
    --names
    --filters.names({  names = names_bought }),
    --filters.names({  names = names_tobe }),
    
    
    
    -------------------------------------------------------------------------------------------------------------
    --codes
    --filters.codes({ codes = codes_tobe })
    
    --------------------------------------------------------------------------------------------------------------
    
    --groups
    --filters.groups( { groups = { "两桶油改革" } } ),

    --------------------------------------------------------------------------------------------------------------
    
    -- 很高的 IO,
    --filters.io({  io_lower = 1.75, io_upper = 100, ch_lower = 1, ch_upper = 11, big_in_lower = 0, date_offset = -1 }),
    
    --------------------------------------------------------------------------------------------------------------

    -- 中高 IO, 高 CH
    --filters.io({  io_lower = 0, io_upper = 100, ch_lower = -1.5, ch_upper = 11, big_in_lower = 0, date_offset = -1 }),
    --filters.io({  io_lower = 1.4, io_upper = 100, ch_lower = 5, ch_upper = 11, big_in_lower = 0, date_offset = 0 }),
    
    --------------------------------------------------------------------------------------------------------------

    -- 中高 IO, 低 CH
    --filters.io({  io_lower = 0, io_upper = 100, ch_lower = -1, ch_upper = 5, big_in_lower = 0, date_offset = -1 }),
    filters.io({  io_lower = 1.4, io_upper = 100, ch_lower = 1, ch_upper = 5, big_in_lower = 0, date_offset = 0}),
    
    
    --------------------------------------------------------------------------------------------------------------
    
    -- 反 IO, 高 CH
    --filters.io({  io_lower = 0.7, io_upper = 1, ch_lower = 3, ch_upper = 11, big_in_lower = 0, date_offset = 0 }),
    
    --------------------------------------------------------------------------------------------------------------
    
    -- 非常低 IO, 正 CH, 蓄力股
    --filters.io({  io_lower = 0.5, io_upper = 1.2, ch_lower = -0.5, ch_upper = 5, big_in_lower = 0, date_offset = -1 }),
    --filters.io({  io_lower = 1, io_upper = 1.5, ch_lower = 5, ch_upper = 8, big_in_lower = 0, date_offset = 0 }),
    
    --------------------------------------------------------------------------------------------------------------
    
    -- ch 0
    --filters.io({  io_lower = 0.8, io_upper = 1.2, ch_lower = -0.1, ch_upper = 0.1, big_in_lower = 0, date_offset = -1 })
    
    ---------------------------------------------------------------------------------------------
    
    -- 迥 CH
    --filters.io({  io_lower = 0, io_upper = 100, ch_lower = 9, ch_upper = 11, big_in_lower = 0, date_offset = 0 }),
    
    --------------------------------------------------------------------------------------------------------------
    
    --anti io series
    --filters.io({  io_lower = 0.8, io_upper = 100, ch_lower = -0.1, ch_upper = 0.15, big_in_lower = 0, date_offset = -1  }),
    --filters.io({  io_lower = 1.3, io_upper = 100, ch_lower = 3, ch_upper = 11, big_in_lower = 0, date_offset = 0  }),
    --filters.io_all({  io_lower = 0, io_upper = 100, ch_lower = -3, ch_upper = 11, big_in_lower = 0, date_offset = 0  }),
    
    --------------------------------------------------------------------------------------------------------------
    -- anti io series 2
    --filters.io({  io_lower = 1.1, io_upper = 100, ch_lower = -1.5, ch_upper = 0, big_in_lower = 0, date_offset = -1  }),
    --filters.io({  io_lower = 0.5, io_upper = 3, ch_lower = 1, ch_upper = 11, big_in_lower = 0, date_offset = 0  }),
    --filters.io_all({  io_lower = 0, io_upper = 100, ch_lower = -3, ch_upper = 11, big_in_lower = 0, date_offset = 0  }),

    
    -- two days anti io io >= 2 plus high io
    --filters.io({  io_lower = 0.6, io_upper = 0.97, ch_lower = -0.1, ch_upper = 3, big_in_lower = 0, date_offset = -1 }),
    --filters.io({  io_lower = 0.6, io_upper = 0.97, ch_lower = -0.1, ch_upper = 3, big_in_lower = 0, date_offset = 0 }),
    
    
    
    --------------------------------------------------------------------------------------------------------------
    
    -- high io, ch == 0
    --filters.io({  io_lower = 1.05, io_upper = 100, ch_lower = -0.1, ch_upper = 0.1, big_in_lower = 0 })
    
    --------------------------------------------------------------------------------------------------------------

    -- serial anti io
    --filters.io({  io_lower = 0.7, io_upper = 1, ch_lower = -0.1, ch_upper = 8, big_in_lower = 0, date_offset = -2 }),
    --filters.io({  io_lower = 0.7, io_upper = 1, ch_lower = -0.1, ch_upper = 8, big_in_lower = 0, date_offset = -1 }),
    -- filters.io({  io_lower = 0.7, io_upper = 1, ch_lower = -0.1, ch_upper = 8, big_in_lower = 0, date_offset = 0 }),
    
    
    
    
    --------------------------------------------------------------------------------------------------------------
    
    -- high io, io >= 1.4 && ch >= 0 && ch <= 2
    --filters.io({  io_lower = 1.4, io_upper = 100, ch_lower = 0, ch_upper = 2, big_in_lower = 0 })
    
    --------------------------------------------------------------------------------------------------------------
    
    -- moderate io and low ch
    --filters.io({  io_lower = 1.35, io_upper = 100, ch_lower = 0, ch_upper = 2, big_in_lower = 0 })
    
    -------------------------------------------------------------------------------------------------------------    

    -- anti io, 0.5 <= io <= 0.98, ch >= 0, ch <= 3
    --filters.io({  io_lower = 0.8, io_upper = 0.98, ch_lower = 1, ch_upper = 3, big_in_lower = 0, date_offset = 0 }),
    
    -------------------------------------------------------------------------------------------------------------    

    --moderate
    --filters.io({  io_lower = 1.2, io_upper = 100, ch_lower = 5, ch_upper = 11, big_in_lower = 0  }),
    
    -- low io
    --filters.io({  io_lower = 0.8, io_upper = 1.25, ch_lower = 0, ch_upper = 5, big_in_lower = 0  })
    
    
    
    
    --very high io
    --filters.io({  io_lower = 1.5, io_upper = 100, ch_lower = -1, ch_upper = 10.5, big_in_lower = 0 }),
    
    -- io ceil
    --filters.io({  io_lower = 0, io_upper = 100, ch_lower = 8, ch_upper = 11, big_in_lower = 0 }),
    
    -- flow in increase
    --filters.io_increase({ in_lower = 30, in_upper = 100, in_swing = 5, ch_lower = -10, ch_upper = 10 })
    
    -- chase high 
    --filters.io({  io_lower = 1.2, io_upper = 100, ch_lower = 5, ch_upper = 10, big_in_lower = 0 }),
    
    -- find underline
    --filters.io_any({  io_lower = 1.85, io_upper = 100, ch_lower = 0, ch_upper = 10, big_in_lower = 0  }),
    
    --filters.io_any({  io_lower = 0.5, io_upper = 0.75, ch_lower = -1, ch_upper = 5, big_in_lower = 0  }),
    --filters.io_any({  io_lower = 1.75, io_upper = 100, ch_lower = 0, ch_upper = 10.5, big_in_lower = 0, ch_avg_lower = 0, ch_avg_upper = 10  }),
    
    --all
    --filters.io_all({  io_lower = 1, io_upper = 100, ch_lower = -2.5, ch_upper = 10.5, big_in_lower = 0, ch_avg_lower = 0, ch_avg_upper = 10  }),
    
    
    
}

th_mod_flow_inst:go(opts)