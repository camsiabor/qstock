if 1 == 1 then
    local t = os.time()
    print(t)
    return
end

local global = require("q.global")



local reqopts = {}
local url = "http://www.aastocks.com/sc/stocks/market/index/h-shares.aspx?t=1&hk=0"
url = "http://nufm.dfcfw.com/EM_Finance2014NumericApplication/JS.aspx?cb=jQuery112409831336260877317_1554968702483&type=CT&token=4f1862fc3b5e77c150a2b985b12db0fd&sty=FCABHL&js=(%7Bdata%3A%5B(x)%5D%2CrecordsFiltered%3A(tot)%7D)&cmd=C._AHH&st=(AB%2FAH%2FHKD)&sr=-1&p=2&ps=20&_=1554968702501"
local reqopt = {}
reqopt["url"] = url
reqopts[1] = reqopt


local err
local browser = global["gorilla"]
reqopts, err = browser.Get(reqopts, 0, 0, 1, 1)
if err ~= nil then
    print(err)
    return
end

reqopt = reqopts[1]

local html = reqopt["content"]

print(html)