local th_mod_fund = require("sync.th.mod_fund")
local th_mod_fund_inst = th_mod_fund:new()

local cache_code = Q.cachem.Get("stock.code");
local codes = cache_code.Get(false, "sh");
local codes_fragment = {}

for i = 1, #codes do
    codes_fragment[#codes_fragment + 1] = codes[i]
end

local data = {}
local result = {}
local opts = {}
opts.loglevel = 0
opts.browser = "wget"
opts.codes = codes_fragment

print(#opts.codes)

opts.concurrent = 20
opts.newsession = false

opts.dofetch = false

opts.db = "flow"
opts.persist = true
opts.print_from = 1
opts.print_to = 1

opts.filter2 = function(opts, data, result)
    local n = #data
    for i = 1, n do
        local one = data[i]
    end
end

opts.print_data2 = function()

end


th_mod_fund_inst:go(opts, data, result)