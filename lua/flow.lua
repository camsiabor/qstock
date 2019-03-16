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

local simple = require("common.simple")
local xml = require("common.xml2lua.xml2lua")
local tree_handler = require("common.xml2lua.tree")


function request(opts, result, retry)
    
    local headers = {}
    headers["Host"] = "data.10jqka.com.cn"
    headers["Referer"] = "http://data.10jqka.com.cn/funds/ggzjl/"
    headers["X-Request-With"] = "XMLHttpRequest"
    headers["Upgrade-Insecure-Requests"] = "1"
    headers["Accept"] = "text/html, */*; q=0.01"
    headers["Accept-Language"] = "zh,zh-TW;q=0.9,en-US;q=0.8,en;q=0.7,zh-CN;q=0.6"
    headers["hexin-v"] = opts.token
    
    local url_prefix = "http://data.10jqka.com.cn/funds/ggzjl/field/"..opts.field.."/order/"..opts.order.."/page/"
    local url_suffix = "/ajax/1"
    
    local count = 1
    local reqopts = {}
    for page = opts.from, opts.to do
        local url = url_prefix..page..url_suffix    
        local reqopt = {}
        reqopt["url"] = url
        reqopt["headers"] = headers
        reqopt["encoding"] = "gbk"
        reqopts[count] = reqopt
        count = count + 1
    end
    count = count - 1

    if opts.concurrent <= 1 then
        if opts.newsession then
            reqopts = Q.selenium.GetEx(reqopts, opts.nice)
        else
            reqopts = Q.selenium.Get(reqopts, opts.nice)
        end
    else
        reqopts = Q.selenium.GetConcurrent(reqopts, opts.nice, opts.concurrent, opts.newsession)
    end
    
    for i = 1, count do
        local reqopt = reqopts[i]
        response_handle(opts, result, reqopt)
    end

end


function response_handle(opts, result, reqopt)
    local url = reqopt["url"]
    local html = reqopt["content"]
    
    if html == nil then
        print("[error] request failure")
        print(url)
        return
    end
    
    local tree = tree_handler:new()
    local parser = xml.parser(tree)
    parser:parse(html)
    
    local htable = tree.root.html.body.table
    if htable == nil then
        print("[error] response content invalid "..#html)
        print(url)
        print(html)
        print("")
        return
    end

    local tbody = htable.tbody
    
    local tr_count = #tbody.tr
    
    for i = 1, tr_count do
        
        local tr = tbody.tr[i]
        local index = tr.td[1][1]
        local code = tr.td[2].a[1]
        local name = tr.td[3].a[1]
        local change_rate = tr.td[5][1]
        local turnover = tr.td[6][1]
        local flow_in = tr.td[7][1]
        local flow_out = tr.td[8][1]
        local flow = tr.td[9][1]
        local amount = tr.td[10][1]
        local flow_big = tr.td[11][1]
        
        turnover = string.gsub(turnover, "%%", "") + 0
        change_rate = string.gsub(change_rate, "%%", "") + 0
        
        flow = simple.str2num(flow)
        flow_in = simple.str2num(flow_in)
        flow_out = simple.str2num(flow_out)
        flow_big = simple.str2num(flow_big)
        
        flow_in = simple.nozero(flow_in)
        flow_out = simple.nozero(flow_out)
        flow_big = simple.nozero(flow_big)
        
        amount = simple.str2num(amount)
        
        local flow_big_rate = flow_big / amount * 100
        local flow_big_rate_compare = flow / flow_big
        local flow_big_rate_total = turnover * flow_big_rate / 100
        
        local flow_in_rate = flow_in / amount * 100
        local flow_out_rate = flow_out / amount * 100
        local flow_io_rate = flow_in / flow_out
        
        local flow_big_in_rate = flow_big / flow_in * 100
        
        --local flow_big_rate_cross = (turnover * amount * flow_big_rate / 100) * flow_io_rate * flow_big_in_rate
        local flow_big_rate_cross = flow_io_rate * flow_big_rate_total * flow_big_rate / 100 * flow_big_in_rate
        local change_rate_ex = change_rate
        if change_rate_ex < 0 then
            change_rate_ex = 0.1
        end
        local flow_big_rate_cross_ex = flow_big_rate_cross / (change_rate_ex + 2.5)

        flow_big_rate = simple.numcon(flow_big_rate)
        flow_big_rate_compare = simple.numcon(flow_big_rate_compare)
        flow_big_rate_total = simple.numcon(flow_big_rate_total)
        flow_big_rate_cross = simple.numcon(flow_big_rate_cross)
        flow_big_rate_cross_ex = simple.numcon(flow_big_rate_cross_ex)
        
        flow_in_rate = simple.numcon(flow_in_rate)
        flow_out_rate = simple.numcon(flow_out_rate)
        flow_io_rate = simple.numcon(flow_io_rate)
        
        flow_big_in_rate = simple.numcon(flow_big_in_rate)
        
        local critical = change_rate >= opts.ch_lower and change_rate <= opts.ch_upper
        if critical then
            critical = flow_big_rate_compare >= opts.big_c_lower and flow_big_rate_compare <= opts.big_c_upper
        end
        
        if critical then
            
            local one = {}
            one.index = index
            one.code = code
            one.name = name
            one.flow = flow
            one.flow_in = flow_in
            one.flow_out = flow_out
            one.amount = amount
            one.turnover = turnover
            one.flow_big = flow_big
            one.change_rate = change_rate
            
            one.flow_big_rate = flow_big_rate
            one.flow_big_rate_total = flow_big_rate_total
            one.flow_big_rate_compare = flow_big_rate_compare
            one.flow_big_rate_cross = flow_big_rate_cross
            one.flow_big_rate_cross_ex = flow_big_rate_cross_ex
            
            one.flow_in_rate = flow_in_rate
            one.flow_out_rate = flow_out_rate
            one.flow_io_rate = flow_io_rate
            
            one.flow_big_in_rate = flow_big_in_rate
            
            result[#result + 1] = one
        
        end -- if ciritical end
        
    end -- for tr end
end

------------------------------------------------------------------------------------------

local opts = {}
local result = {}

opts.from = 1
opts.to = 20
opts.nice = 1000
opts.concurrent = 0
opts.newsession = false

opts.ch_lower = -2.5
opts.ch_upper = 6
opts.big_c_lower = 0.2
opts.big_c_upper = 10

opts.field = "zjjlr"
opts.order = "desc"

request(opts, result)
    

simple.table_sort(result, "flow_big_rate_cross_ex")

local print_head = "i\tcode\tname\tch\tturn\tio\tin\tbig_in\tbig_r\tbig_t\tbig_c\tcross\tcross2\tbig"
local count = 1
for i = 1, #result do
    local one = result[i]
    if count % 10 == 1 then
        print("")
        print(print_head)
    end
    print(one.index.."\t"..one.code.."\t"..one.name.."\t"..one.change_rate.."\t"..one.turnover.."\t"..one.flow_io_rate.."\t"..one.flow_in_rate.."\t"..one.flow_big_in_rate.."\t"..one.flow_big_rate.."\t"..one.flow_big_rate_total.."\t"..one.flow_big_rate_compare.."\t"..one.flow_big_rate_cross.."\t"..one.flow_big_rate_cross_ex.."\t"..one.flow_big)
    count = count + 1
end