--local mobdebug = require("mobdebug")
--mobdebug.start()




local http = require("socket.http")
local url = "http://www.bing.cn"
url = "http://data.10jqka.com.cn/funds/ggzjl/field/zjjlr/order/desc/page/2/ajax/1/"


local headers = {}
headers["hexin-v"] = "AqNsoKE1-wN1F7c32WAV9OtUMuxOmDPfcSd7bdUE_cV4m80aXWjHKoH8C1Hm"
headers["Host"] = "data.10jqka.com.cn"
headers["Referer"] = "http://data.10jqka.com.cn/funds/ggzjl/"
local html = Q.http.Get(url, headers, "gbk")


local xml = require("common.xml2lua.xml2lua")
local tree = require("common.xml2lua.tree")
local parser = xml.parser(tree)
parser:parse(html)

for k, v in pairs(tree.root) do
    print(k)
end


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