-- "http://data.10jqka.com.cn/funds/ggzjl/field/zjjlr/order/desc/page/1/ajax/1/"





--print(html)

local xml = require("common.xml2lua.xml2lua")
local tree = require("common.xml2lua.tree")
local parser = xml.parser(tree)
parser:parse(html)

local tbody = tree.root.table.tbody
local tr_count = #tbody.tr

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


function request(page) 

local headers = {}
headers["hexin-v"] = "AqNsoKE1-wN1F7c32WAV9OtUMuxOmDPfcSd7bdUE_cV4m80aXWjHKoH8C1Hm"
headers["Host"] = "data.10jqka.com.cn"
headers["Referer"] = "http://data.10jqka.com.cn/funds/ggzjl/"
local html = Q.http.Get(url, headers, "gbk")
local url_prefix = "http://data.10jqka.com.cn/funds/ggzjl/field/zjjlr/order/desc/page/"
local url_suffix = "/ajax/1/"

local n
for i = 1, tr_count do
    local tr = tbody.tr[i]
    local code = tr.td[2].a[1]
    local name = tr.td[3].a[1]
    local change_rate = tr.td[5][1]
    local turnover = tr.td[6][1]
    local amount = tr.td[10][1]
    local flow_big = tr.td[11][1]
    
    turnover = string.gsub(turnover, "%%", "") + 0
    change_rate = string.gsub(change_rate, "%%", "") + 0
    
    n = string.find(amount, "亿" )
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
    
    
    print(code.." "..name.." "..change_rate.." "..amount.." "..flow_big)
end
end


--[[
for i, tr in pairs(tbody.tr) do
  print(tr.td)
end
]]--



--print(tree)


--local b, c, h = http.request(url)

--[[
print(Q.http ~= nil)

local api = Q.global.Config["api"]
local tushare = api["tushare"]["profiles"]
local khistory = tushare["k.history"]
local d = khistory["nice"]

print(d)
print(type(d))

return a
]]--

--print(b)

return 0