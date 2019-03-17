

local mod_th_fund = require("sync.th.fund")
local inst = mod_th_fund:new()

local data = {}
local result = {}

local opts = {}
opts.browser = "wget"
opts.codes = { "603178", "000001", "000009" }

opts.concurrent = 1
opts.newsession = false

opts.dofetch = false

opts.db = "flow"
opts.persist = true

inst:go(opts, data, result)
