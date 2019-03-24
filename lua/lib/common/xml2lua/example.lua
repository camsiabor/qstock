


if self.xml == nil then
    self.xml = require("common.xml2lua.xml2lua")
    self.xml_tree_handler = require("common.xml2lua.tree")
end

local html = "<html><h1>hello world</h1></html>"
local tree = self.xml_tree_handler:new()
local parser = self.xml.parser(tree)
parser:parse(html)