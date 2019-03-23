



local M = __q_global
if M == nil then
    print("__q_global is null?")
    return
end

M.__index = M
setmetatable(M, M)


return M