local t = { a = 1 }

local browser_type = "gorilla"

local browser = Q[browser_type]

local html = browser.Get("http://www.baidu.com", nil, "utf-8")

print(html)