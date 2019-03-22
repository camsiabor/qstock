

local M = {}

M.__index = M
M.logger = Q.logger
M.loggerm = Q.loggerm
M.Level = Q.loggerm.LevelInt("info")

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

function M:log(...)
    local debuginfo = debug.getinfo(2, "Snl")
    if debuginfo.name == nil then
        debuginfo.name = debuginfo.what
    end
    self.logger.LogEx(self.Level, -1, debuginfo.short_src, debuginfo.currentline, debuginfo.name,  ...)
end

return M