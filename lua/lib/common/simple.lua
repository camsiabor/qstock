local simple = {}

function simple.is(b)
    local t = type(b)
    if t == "nil" then
        return false
    end
    if t == "boolean" then
        return t
    end
    if t == "number" then
        return t ~= 0
    end
    if t == "string" then
        if t == "" or t == "0" or t == "false" then
            return false
        end
        if t == "true" then
            return true
        end
        return false
    end
    if t == "function" then
        return t()
    end
    return false
end

function simple.path(t, ...)
    local v = t
    local n = select('#', ...)
    for i = 1, n do
        local field = select(i, ...)
        v = v[field]
        if v == nil then
            return nil
        end
    end
    return v
end

function simple.pathex(t, fields)
    local v = t
    local n = #fields
    for i = 1, n do
        local field = fields[i]
        v = v[field]
        if v == nil then
            return nil
        end
    end
    return v
end

function simple.numcon(num, limit)
    if limit == nil then
        limit = 5
    end
    if num >= -0.0001 and num <= 0.0001 then
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

function simple.percent2num(percentstr)
    percentstr = string.gsub(percentstr, "%%", "")
    if percentstr == "" then
        return 0
    end
    if percentstr == "-" or percentstr == "+"  then
        return 0
    end
    return percentstr + 0
end

function simple.table_clone(t)
    if t == nil then
        return nil
    end
    local clone = {}
    for k, v in pairs(t) do
        clone[k] = v
    end
    return clone
end

function simple.table_sort(t, field)
    local single = type(field) ~= "table"
    local n = #t
    local a, b, av, bv
    for i = 1, n do
        for j = 1, n - i do
            a = t[j]
            b = t[j + 1]
            if single then
                av = a[field]
                bv = b[field]
            else
                av = simple.pathex(a, field)
                bv = simple.pathex(b, field)
            end
            if av < bv then
                t[j] = b
                t[j + 1] = a
            end
        end
    end
end



function simple.table_merge(tables)
    local result = {}
    for i = 1, #tables do
        local table = tables[i]
        if table ~= nil then
            for k , v in pairs(table) do
                result[k] = v
            end
        end
    end
    return result

end




function simple.table_print_all(obj)
    if obj == nil then
        return
    end
    for k, v in pairs(obj) do
        print(k, v)
    end
end

function simple.table_count(obj)
    local n = 0
    for _ in pairs(obj) do
        n = n + 1
    end
    return n
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


function simple.metatable_print_all(obj)
    local meta = getmetatable(obj)
    if meta == nil then
        return
    end
    if meta.__index == nil then
        print("[meta]")
    else
        print("[meta] __index")
        meta = meta.__index
    end
    return simple.table_print_all(meta)
end

------------------------------------------------------------------------------------------------------------------------

function simple.func_call(func, ...)
    if func == nil then
        return nil
    end
    if type(func) == "function" then
        return func(...)
    end
    return nil
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

function simple.table_array_print_with_header(array, from, to, fields, headers, header_interval, delimiter, suffix, formatters, key)

    if from <= 0 or from > to then
        return
    end

    if delimiter == nil then
        delimiter = "\n"
    end

    local headstr = ""
    for i = 1, #headers do
        headstr = headstr .. headers[i] .. "\t"
    end

    local fieldsimple = {}
    for i = 1, #fields do
        local field = fields[i]
        fieldsimple[i] = type(field) ~= "table"
    end

    local ikey = 1
    local nfields = #fields
    local key_value_previous = ""
    local header_interval_original = header_interval
    for a = from, to do

        local obj = array[a]
        local print_header = false
        if key == nil then
            if header_interval > 0 then
                if a % header_interval == 1 then
                    print_header = true
                    header_interval = header_interval_original
                end
            end
        else
            local key_value = obj[key]
            print_header = (key_value ~= key_value_previous)
            key_value_previous = key_value
            ikey = ikey + 1
        end

        if print_header then
            printex(delimiter)
            printex(headstr)
            printex(delimiter)
            ikey = 1
        end

        for f = 1, nfields do
            local field = fields[f]
            local easy = fieldsimple[f]
            local v
            if easy then
                v = obj[field]
            else
                v = simple.pathex(obj, field)
            end
            if v == nil then
                v = ""
            else
                if formatters ~= nil then
                    local formatter = formatters[field]
                    if formatter ~= nil then
                        v = formatter(obj, field, v, ikey)
                    end
                end
            end
            printex(v, "")
        end
        printex(delimiter)
    end

    if suffix ~= nil then
        printex(suffix)
    end

end

function simple.map_to_array(map)
    local i = 1
    local array = {}
    for _, v in pairs(map) do
        if v ~= nil then
            array[i] = v
            i = i + 1
        end
    end
    return array
end


function simple.array_to_map(array, field)
    local map = {}
    local n = #array
    for i = 1, n do
        local one = array[i]
        if one ~= nil then
            if field == nil then
                map[one] = one
            else
                local key = one[field]
                if key ~= nil then
                    map[key] = one
                end
            end
        end
    end
    return map
end

function simple.array_intermix(arrays)
    local n = #arrays[1]
    local arraycount = #arrays
    local result = {}
    for i = 1, n do
        for a = 1, arraycount do
            local array = arrays[a]
            local one = array[i]
            if one ~= nil then
                result[#result + 1] = one
            end
        end
    end
    return result
end

function simple.array_append(array, subarray)
    if subarray == nil then
        return array
    end
    local n = #subarray
    for i = 1, n do
        local one = subarray[i]
        if one ~= nil then
            array[#array + 1] = one
        end
    end
    return array
end

function simple.maps_intersect(maps, callback, ...)
    local n = 0
    local m1 = maps[1]
    local mcount = #maps
    for k, v1 in pairs(m1) do

        local intersect = true
        for i = 2, mcount do
            local m = maps[i]
            local vn = m[k]
            if vn == nil then
                intersect = false
                break
            end
        end

        if intersect then
            n = n + 1
            callback(maps, k, n)
        end
    end
    return n
end


function simple.def(t, field, defvalue)
    local currvalue = t[field]
    if currvalue == nil then
        t[field] = defvalue
        return defvalue
    end
    return currvalue
end

function simple.get(t, field, defvalue)
    if t == nil then
        return defvalue
    end
    local v = t[field]
    if v == nil then
        return defvalue
    end
    return v
end


function simple.now_date(formatter)
    if formatter == nil or #formatter == 0 then
        formatter = "%Y%m%d %H%M"
    end
    local t = os.time()
    return os.date(formatter, t)
end

function simple.now_datetime(formatter)
    if formatter == nil or #formatter == 0 then
        formatter = "%Y%m%d %H%M%S"
    end
    local t = os.time()
    return os.date(formatter, t)
end

function simple.now_hour_min(formatter)
    if formatter == nil or #formatter == 0 then
        formatter = "%H%M"
    end
    local t = os.time()
    return os.date(formatter, t)
end

function simple.intstr(n)
    local s = n ..""

    local index = string.find(s, "%.")
    if index == nil then
        return s
    end
    if index == 0 then
        return "0"
    end
    s = string.sub(s, 1, index - 1)
    return s
end

function simple.num_array_align(array, num)
    num = num + 0
    local n = #array - 1
    for i = 1, n do
        local a = array[i] + 0
        local b = array[i + 1] + 0
        if num >= a and num <= b then
            local a_delta = num - a
            local b_delta = b - num
            if a_delta < b_delta then
                return array[i]
            else
                return array[i + 1]
            end
        end
    end
    return nil
end



return simple
