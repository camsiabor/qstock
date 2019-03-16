--local d = os.date("%Y%m%d", os.time())

local json = require('common.json')
local jsonstr = json.encode({ 1, 2, 'fred', {first='mars',second='venus',third='earth'} })
print(jsonstr)




local dates = Q.calendar.List(0, 0, 0, true)

--date.src.field.order.page
local d = dates[1]
local datasrc = "th"
local field = "zjjlr"
local order = "desc"
local page = 1

local key = string.format("%s.%s.%s.%d", datasrc, field, order, page)

local db = "flow"
local dao = Q.daom.Get("main")

_, err = dao.Update(db, d, key, "helloworld", true, 0, nil)
if err == nil then
    print("[persist]", d, key)
else
    print(err)
end

local html, _ = dao.Get(db, d, key, 0, nil)
print(html)


--print(Q.daom.Get(""))