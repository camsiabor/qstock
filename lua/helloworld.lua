local global = require("q.global")
local json = require("common.json")
local simple = require("common.simple")
 
local TOKEN_PERSIST_LIST = "ch.stock.group.concept"
 
local opts = {}
local db = opts.db
if db == nil then
    db = "group"
end
local dao = global.daom.Get("main")
print("[reload] stock group concept")
local map, err = dao.Get(db, "", TOKEN_PERSIST_LIST, 1, nil)
if err ~= nil then
    print("[reload] failure", db, TOKEN_PERSIST_LIST, err)
    return
end
if map == nil or simple.table_count(map) == 0 then
    print("[reload] empty", db, TOKEN_PERSIST_LIST)
    return
end
local groups = {}
for code in pairs(map) do
    local groupstr = map[code]
    if #groupstr > 0 then
        local group = json.decode(groupstr)
        groups[code] = group
    end
end
return groups