local th_mod_fund = require("sync.th.mod_fund")
local th_mod_fund_inst = th_mod_fund:new()


local cache_code = Q.cachem.Get("stock.code");
local codes = cache_code.Get(false, "sz");

-------------------------------------------------------------------------------------

local fetch_from = 1
local fetch_to = 1
local fetch_each = 20

local data = {}
local result = {}
local opts = {}
opts.loglevel = 0
opts.browser = "gorilla"
--opts.browser = "chrome"
--opts.browser = "firefox"

opts.concurrent = fetch_each
opts.newsession = false

opts.dofetch = true

opts.db = "flow"
opts.persist = true
opts.print_data_from = 1
opts.print_data_to = 1

opts.find_not_curr = true
opts.codes_not_curr = {}

---------------------------------------------------------------------------------------------

if opts.find_not_curr then
    opts.codes = codes
    opts.dofetch = false
    th_mod_fund_inst:go(opts, data, result)
    
    opts.fid_not_curr = false
    local n = #opts.codes_not_curr
    print("codes not current count", n)
    if n > 0 then
        
        -- refetch not current data
        for i = 1, n do
            codes = opts.codes_not_curr
        end
        opts.data = {}
        opts.result = {}
        opts.dofetch = true
    end
end

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