-- http://data.10jqka.com.cn/funds/ggzjl/
-- http://data.10jqka.com.cn/funds/ggzjl/field/zjjlr/order/desc/page/1/ajax/1/

local simple = require("common.simple")
local mod_th_flow = require("sync.th.flow")
local th_flow = mod_th_flow:new()

local opts_current = {}

opts_current.data = {}
opts_current.result = {}

opts_current.debug = false
opts_current.loglevel = 0
opts_current.browser = "firefox"

opts_current.from = 1
opts_current.to = 71
opts_current.nice = 0
opts_current.concurrent = 1
opts_current.newsession = false
opts_current.persist = true

opts_current.dofetch = false
opts_current.date_offset = -1

opts_current.pagesize = 71
opts_current.ch_lower = -2.5
opts_current.ch_upper = 6
opts_current.big_c_lower = 0.2
opts_current.big_c_upper = 10

opts_current.db = "flow"
opts_current.datasrc = "th"
opts_current.field = "zjjlr"
opts_current.order = "desc"

opts_current.sort_field = "flow_big_rate_cross_ex"

opts_current.filter_balance_io_rate = function(opts, data, result)
    print("[filter] low")
    local n = #data
    for i = 1, n do
        local one = data[i]
        local critical =
        one.flow_io_rate >= 0.9 and one.flow_io_rate <= 1.1
                and one.flow_big_in_rate >= 10
                and one.change_rate >= -1.5 and one.change_rate <= 1.5

        if critical then
            result[#result + 1] = one
        end
    end
end

--opts.filter = opts.filter_balance_io_rate

th_flow:go(opts_current)
---------------------------------------------------------------------------------------------------

local opts_prev = simple.table_clone(opts_current)
opts_prev.data = {}
opts_prev.result = {}
opts_prev.dofetch = false
opts_prev.date_offset = -1

th_flow:go(opts_prev)


