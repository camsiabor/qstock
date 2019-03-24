


local global = require("q.global")

local M = {}

M.__index = M
M.logger = global.logger
M.loggerm = global.loggerm
M.Level = global.loggerm.LevelInt("info")

M.opts_def = {
    key = "lua",
    stdout = true,
    dir = "log",
    prefix = "lua",
    suffix = ".log",
}

function M:new(opts)
    local inst = {}
    inst.__index = self
    setmetatable(inst, self)
    if opts == nil then
        opts = {}
    end
    if opts.logger == nil then
        if opts.suffix ~= nil and opts.prefix ~= nil then
            M.loggerm.New(opts.key, opts.dir, opts.prefix, opts.suffix, opts.stdout)
        end
    else
        inst.logger = opts.logger
    end
    return inst
end

function M:log(level, ...)
    if level == nil then
        level = self.level
        if level == nil then
            level = 2
        end
    end
    local debuginfo = debug.getinfo(2, "Snl")
    if debuginfo.name == nil then
        debuginfo.name = debuginfo.what
    end
    self.logger.LogEx(level, -1, debuginfo.short_src, debuginfo.currentline, debuginfo.name,  ...)
end


function M:debug(...)
    self:log(0, ...)
end

function M:verbose(...)
    self:log(1, ...)
end

function M:info(...)
    self:log(2, ...)
end

function M:warn(...)
    self:log(3, ...)
end

function M:error(...)
    self:log(4, ...)
end

function M:fatal(...)
    self:log(5, ...)
end


function M:trace(...)
    local stack = ""
    for i = 2, 64 do
        local debuginfo = debug.getinfo(i, "Snl")
        if debuginfo == nil then
            break
        end
        if debuginfo.name == nil then
            debuginfo.name = debuginfo.what
        end
        local one = "\t" .. debuginfo.short_src .. " " ..  debuginfo.currentline .. " " .. debuginfo.name
        stack = stack .. one .. "\n"
    end
    self.logger.LogEx(self.Level, -1, ..., "\n", stack)
end

return M