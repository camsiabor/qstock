

--local p = "../../src/github.com/camsiabor/qstock/lua/?.lua"
--local m_package_path = package.path
--package.path = string.format("%s;%s?.lua;%s?/init.lua", m_package_path, p, p)

--require("common.util")

print("[lua] i am test")

require("helloworld")
require("common.util")


print(LUA_PATH)
print(package.path)
print(package.cpath)

-- local x = add(10,20)
x = 1
return x