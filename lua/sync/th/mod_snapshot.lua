
local M = {}
M.__index = M

local global = require("q.global")
local json = require('common.json')
local simple = require("common.simple")

-------------------------------------------------------------------------------------------

function M:new()
    local inst = {}
    inst.__index = self
    setmetatable(inst, self)
    return inst
end

-------------------------------------------------------------------------------------------

function M:snapshots(codes)
    local map = {}
    local n = #codes
    for i = 1, n do
        local code = codes[i]

    end
    return map
end

return M