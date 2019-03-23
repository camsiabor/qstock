local opts = {}
local result = {}
opts.debug = false
opts.ch_lower = -2.5
opts.ch_upper = 6
opts.big_c_lower = 0.2
opts.big_c_upper = 10

opts.field = "zjjlr"
opts.order = "desc"
opts.token = "AiPsICG1e4E4dze1mhyVdGvUsmzOGLlE8dD7mFWIfHn4PE2a3ehHqgF8i6pm"

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
local reqopts = {}
local from = 1
local to = 8
local n = 1
for i = from, to do
    local url = url_prefix..i..url_suffix
    local one = {}
    one["url"] = url
    one["headers"] = headers
    one["encoding"] = "gbk"
    reqopts[n] = one
    n = n + 1
end
n = n - 1

reqopts = global.http.Gets(reqopts)

for i = 1, n do

    local one = reqopts[i]
    local url = one["url"]
    local html = one["content"]
    print("")
    print(url)
    print(#html)
end