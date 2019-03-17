local simple = {}

function simple.str2num(str, keep)
    if keep == nil then
        keep = 5
    end
    local n = string.find(str, "亿")
    if n == nil then
        str = string.gsub(str, "万", "") + 0
        str = str / 10000
    else
        str = string.gsub(str, "亿", "") + 0
    end
    return string.sub(str, 1, keep) + 0
end

function simple.numcon(num, limit)
    if limit == nil then
        limit = 5
    end
    if num < 0.0001 then
        num = 0
    end
    num = num .. ""
    num = string.sub(num, 1, limit)
    return num + 0
end

function simple.nozero(num)
    if num <= 0 then
        num = 0.0001
    end
    return num
end

function simple.percent2num(pencentstr)
    return string.gsub(pencentstr, "%%", "") + 0
end

function simple.table_sort(t, field)
    local n = #t
    for i = 1, n do
        for j = 1, n - i do
            local a = t[j]
            local b = t[j + 1]
            if a[field] < b[field] then
                t[j] = b
                t[j + 1] = a
            end
        end
    end
end

function simple.table_print(obj, fields, suffix)
    local n = #fields
    for i = 1, n do
        local field = fields[i]
        local v = obj[field]
        printex(v, "")
    end
    if suffix ~= nil then
        printex(suffix)
    end
end

function simple.table_array_print(array, fields, delimiter, suffix)
    if delimiter == nil then
        delimiter = "\n"
    end
    local narray = #array
    local nfields = #fields
    for a = 1, narray do
        local obj = array[a]
        for f = 1, nfields do
            local field = fields[f]
            local v = obj[field]
            printex(v, "")
        end
        printex(delimiter)
    end
    if suffix ~= nil then
        printex(suffix)
    end
end

function simple.table_array_print_with_header(array, from, to, fields, headers, header_interval, delimiter, suffix)
    if delimiter == nil then
        delimiter = "\n"
    end

    local headstr = ""
    for i = 1, #headers do
        headstr = headstr .. headers[i] .. "\t"
    end

    local nfields = #fields
    local header_interval_original = header_interval
    for a = from, to do

        if header_interval > 0 then
            if a % header_interval == 1 then
                printex(delimiter)
                printex(headstr)
                printex(delimiter)
                header_interval = header_interval_original
            end
        end

        local obj = array[a]
        for f = 1, nfields do
            local field = fields[f]
            local v = obj[field]
            printex(v, "")
        end
        printex(delimiter)
    end

    if suffix ~= nil then
        printex(suffix)
    end

end

return simple
