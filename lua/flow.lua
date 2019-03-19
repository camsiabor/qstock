local th_mod_fund = require("sync.th.mod_fund")
local th_mod_fund_inst = th_mod_fund:new()


local cache_code = Q.cachem.Get("stock.code");
local codes = cache_code.Get(false, "sz");

-------------------------------------------------------------------------------------

local fetch_from = 1
local fetch_to = 0
local fetch_each = 50

local data = {}
local result = {}
local opts = {}
opts.loglevel = 0
opts.browser = "wget"
--opts.browser = "chrome"

opts.concurrent = 10
opts.newsession = false

opts.dofetch = false

opts.db = "flow"
opts.persist = true
opts.print_data_from = 1
opts.print_data_to = 1

---------------------------------------------------------------------------------------------



local fragment = {}
if fetch_to <= 0 then
    fetch_to = #codes
end
for i = fetch_from, fetch_to do
    local code = codes[i]
    fragment[#fragment + 1] = code
    if i == fetch_to or i % fetch_each == 0 then
        opts.i = i
        opts.codes = fragment
        th_mod_fund_inst:go(opts, data, result)
        fragment = {}
    end
end

print("[fetch] from " .. fetch_from .. " to " .. fetch_to)