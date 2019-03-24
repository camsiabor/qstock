if self.htmlparser == nil then
    self.htmlparser = require("common.htmlparser.htmlparser")
end

local root = self.htmlparser.parse(html_table)
local tbody = root:select("tbody")[1]
local tr_count = #tbody.nodes
for i = 1, tr_count do
    local tr = tbody.nodes[i]
    local tds = tr.nodes
    local td = tds[1]
    print(td:getcontent())
    print(simple.metatable_print_all(td))
    print("----------------------------------")
end -- for tr end