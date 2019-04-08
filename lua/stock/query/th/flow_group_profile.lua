local global = require("q.global")

local th_mod_flow = require("sync.th.mod_flow")

local th_mod_flow_inst = th_mod_flow:new()

local opts = {}

opts.debug = false
opts.loglevel = 0
opts.browser = "firefox"

opts.request = false
opts.request_from = 1
opts.request_to = 71
opts.nice = 0

opts.concurrent = 5
opts.newsession = true
opts.persist = true

opts.date_show = 10

opts.date_offset = 0
opts.date_offset_to = 0
opts.date_offset_from = -10

opts.db = "flow"
opts.datasrc = "th"
opts.field = "zjjlr"
opts.order = "desc"

opts.link_stock_group = true
opts.link_stock_snapshot = false

opts.sort_field = "avg_io"

th_mod_flow_inst:go_stock_group_profile(opts)