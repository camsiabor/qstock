local url = "http://q.10jqka.com.cn/gn/detail/code/301558/"

local opts = {}
opts.browser = "gorilla"

local reqopts = {}
local reqopt = {}
reqopt["url"] = url
reqopt["encoding"] = "gbk"
reqopts[1] = reqopt

local err
local browser = Q[opts.browser]
reqopts, err = browser.Get(reqopts, 0, false, 1, 0)
if err ~= nil then
    print("[list] [request] fatal", err)
    return
end

local html = reqopts[1]["content"]
--print(html)

local tag_table_start = '<table class="m%-table m%-pager%-table">'
local tag_table_end = '</table>'

local index = string.find(html, tag_table_start)
if index == nil then
    print("[list] [request] failure")
    print(html)
    return
end
local html_table = string.sub(html, index)
local index_table_end = string.find(html_table, tag_table_end)
html_table = string.sub(html_table, 1, index_table_end + #tag_table_end)


print(#html_table)

local tag_page_info_start = '<span class="page_info">'
local tag_page_info_end = '</span>'
local index_tag_page_info_start = string.find(html, tag_page_info_start, index_table_end + #tag_table_end + 1)
local index_tag_page_info_end = string.find(html, tag_page_info_end, index_tag_page_info_start + #tag_page_info_start + 1)
local html_page_info = string.sub(html, index_tag_page_info_start + #tag_page_info_start + 2, index_tag_page_info_end - 1)
print(html_page_info)