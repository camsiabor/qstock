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
    headers["hexin-v"] = "AlSbNepg9I5NFWDCq81qCZDZJZnFrXk-utYMye404s9jQvoPFr1IJwrh3G49"
    headers["Host"] = "data.10jqka.com.cn"
    headers["Referer"] = "http://data.10jqka.com.cn/funds/ggzjl/"
    
    local url_prefix = "http://data.10jqka.com.cn/funds/ggzjl/field/zjjlr/order/desc/page/"
    local url_suffix = "/ajax/1/"
    local url = url_prefix..page..url_suffix
    
    local html, response_header = Q.http.Get(url, headers, "gbk")
    

    local tree = tree_handler:new()
    local parser = xml.parser(tree)
    parser:parse(html)


    local tbody = tree.root.table.tbody
    local tr_count = #tbody.tr
    
    for i = 1, tr_count do
        
        
        local tr = tbody.tr[i]
        local index = tr.td[1][1]
        local code = tr.td[2].a[1]
        local name = tr.td[3].a[1]
        local change_rate = tr.td[5][1]
        local turnover = tr.td[6][1]
        local amount = tr.td[10][1]
        local flow_big = tr.td[11][1]
        
        turnover = string.gsub(turnover, "%%", "") + 0
        change_rate = string.gsub(change_rate, "%%", "") + 0
        
        local critical = change_rate >= opts.ch_lower and change_rate <= opts.ch_upper
        
        
        if critical then
            local n = string.find(amount, "亿" )
            if n == nil then
                amount = string.gsub(amount, "万", "") + 0
                amount = amount / 10000
            else
                amount = string.gsub(amount, "亿", "") + 0
            end
    
            n = string.find(flow_big, "亿" )
            if n == nil then
                flow_big = string.gsub(flow_big, "万", "") + 0
                flow_big = flow_big / 10000
            else
                flow_big = string.gsub(flow_big, "亿", "") + 0
            end
            
            local one = {}
            one.index = index
            one.code = code
            one.name = name
            one.amount = amount
            one.turnover = turnover
            one.flow_big = flow_big
            one.change_rate = change_rate
            
            result[#result + 1] = one
            
            print(index.." "..code.." "..name.." "..change_rate.." "..amount.." "..flow_big)
        end
            
        
        
    end
end

local opts = {}
local result = {}
opts.ch_lower = -2
opts.ch_upper = 5

request(4, opts, result)
request(1, opts, result)

return 0