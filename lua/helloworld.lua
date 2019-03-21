local browser_type = "gorilla"

local browser = Q[browser_type]

local opts = {}
for i = 1, 10 do 
    local opt = {}
    opt["url"] = "http://www.baidu.com"
    opts[#opts + 1] = opt
end

opts = browser.Get(opts, 0, false, 0, 0)

for i = 1, #opts do 
    local opt = opts[i]
    local html = opt["content"]
    print(#html)
end