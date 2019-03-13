-- http://data.10jqka.com.cn/funds/ggzjl/
-- http://data.10jqka.com.cn/funds/ggzjl/field/zjjlr/order/desc/page/1/ajax/1/



--[[ 
     1序号	
    2股票代码	
    3股票简称	
    4最新价	
    5涨跌幅	
    6换手率	
    7流入资金(元)	
    8流出资金(元)	
    9净额(元)	
    10成交额(元)	
    11大单流入(元)
]]--

local xml = require("common.xml2lua.xml2lua")
local tree_handler = require("common.xml2lua.tree")

function request(page, opts, result)

    local headers = {}
    headers["hexin-v"] = "Ane4LC2R58dSPWNL9Hd5sB-4BmDCPEEnJQrPjskskW2vLpmW0Qzb7jXgX3Ta"
    headers["Host"] = "data.10jqka.com.cn"
    headers["Referer"] = "http://data.10jqka.com.cn/funds/ggzjl/"
    headers["X-Request-With"] = "XMLHttpRequest"
    
    local url_prefix = "http://data.10jqka.com.cn/funds/ggzjl/field/zjjlr/order/desc/page/"
    local url_suffix = "/ajax/1/"
    local url = url_prefix..page..url_suffix
    
    local html = Q.http.Get(url, headers, "gbk")

    local tree = tree_handler:new()
    local parser = xml.parser(tree)
    parser:parse(html)

    if tree.root.table == nil then
        print("[error] request failure")
        print(url)
        return
    end

    local tbody = tree.root.table.tbody
    
    local tr_count = #tbody.tr
    
    for i = 1, tr_count do
        
        
        local tr = tbody.tr[i]
        local index = tr.td[1][1]
        local code = tr.td[2].a[1]
        local name = tr.td[3].a[1]
        local change_rate = tr.td[5][1]
        local turnover = tr.td[6][1]
        local flow = tr.td[9][1]
        local amount = tr.td[10][1]
        local flow_big = tr.td[11][1]
        
        turnover = string.gsub(turnover, "%%", "") + 0
        change_rate = string.gsub(change_rate, "%%", "") + 0
        
        local n = string.find(flow, "亿" )
        if n == nil then
            flow = string.gsub(flow, "万", "")
            flow = flow / 10000
        else
            flow = string.gsub(flow, "亿", "")
        end
        flow = string.sub(flow, 1, 5) + 0
        
        n = string.find(amount, "亿" )
        if n == nil then
            amount = string.gsub(amount, "万", "")
            amount = amount / 10000
        else
            amount = string.gsub(amount, "亿", "")
        end
        amount = string.sub(amount, 1, 5) + 0
        

        n = string.find(flow_big, "亿" )
        if n == nil then
            flow_big = string.gsub(flow_big, "万", "")
            flow_big = flow_big / 10000
        else
            flow_big = string.gsub(flow_big, "亿", "")
        end
        flow_big = string.sub(flow_big, 1, 5) + 0
        
        local flow_big_rate = flow_big / amount * 100
        local flow_big_rate_compare = flow / flow_big
        local flow_big_rate_total = turnover * flow_big_rate / 100
        
        flow_big_rate = string.sub(flow_big_rate.."", 1, 5) + 0
        flow_big_rate_compare = string.sub(flow_big_rate_compare.."", 1, 5) + 0
        flow_big_rate_total = string.sub(flow_big_rate_total.."", 1, 5) + 0
        
        
        local critical = change_rate >= opts.ch_lower and change_rate <= opts.ch_upper
        
        if critical then
            
            local one = {}
            one.index = index
            one.code = code
            one.name = name
            one.flow = flow
            one.amount = amount
            one.turnover = turnover
            one.flow_big = flow_big
            one.change_rate = change_rate
            
            one.flow_big_rate = flow_big_rate
            one.flow_big_rate_total = flow_big_rate_total
            one.flow_big_rate_compare = flow_big_rate_compare
            
            result[#result + 1] = one
            
            --print(index.."\t"..code.."\t"..name.."\t"..change_rate.."\t"..amount.."\t"..flow_big)
        end -- ciritical end
    end -- for tr end
    
end

local opts = {}
local result = {}
opts.ch_lower = -2
opts.ch_upper = 3


for i = 1, 4 do
    request(i, opts, result)
end

local n = #result
for i = 1, n do
    for j = 1, n - i do
        local a = result[j]
        local b = result[j + 1]
        if a.flow_big_rate_total < b.flow_big_rate_total then
            result[j] = b
            result[j + 1] = a
        end
    end
end

local print_head = "i\tcode\tname\tch\tbig_r\tbig_T\tbig_c\tbig"
print(print_head)
for i = 1, #result do
    local one = result[i]
    
    if i % 10 == 0 then
        print(print_head)
    end
    
    print(one.index.."\t"..one.code.."\t"..one.name.."\t"..one.change_rate.."\t"..one.flow_big_rate.."\t"..one.flow_big_rate_total.."\t"..one.flow_big_rate_compare.."\t"..one.flow_big)
    
    
end

--request(1, opts, result)

return 0