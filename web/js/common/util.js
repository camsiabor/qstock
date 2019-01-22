
QUtil.DAY_MILLI = 24 * 3600 * 1000;



QUtil.COLOR_UP = '#f04864';
QUtil.COLOR_EVEN = "#888888";
QUtil.COLOR_DOWN =  '#2fc25b';


function QUtil(opt) {
    opt = opt || {};
    for(let k in opt) {
        this[k] = opt[k];
    }
    for(let k in QUtil) {
        this[k] = QUtil[k];
    }
    this.context = this.context || window;
}

QUtil.get = function(obj, pathes, defval) {
    let current = obj;
    try {
        for (let i = 0; i < pathes.length; i++) {
            let path = pathes[i];
            current = current[path];
            if (!current) {
                return defval;
            }
        }
    } catch (ex) {
        return defval;
    }
    return current;
};

QUtil.prototype.handle_error = function(error) {
    let err = error.stack || error;
    if (this.context && this.context.console) {
        this.context.console.text = err;
    }
    window.console.error(err);
};

QUtil.prototype.handle_response = function(resp, printer, msg) {

    if (!resp) {
        return resp;
    }

    if (resp instanceof Array) {
        return resp;
    }

    let r = resp.data;

    if (typeof r === "string") {
        if (r.charAt(0) !== '{') {
            // let zbinary = Base64.decode(r);
            let zbinary = Base64.decodeToBytes(r);///important
            let uarr = new Uint8Array(new ArrayBuffer(zbinary.length));
            for (let i = 0, n = zbinary.length; i < n; ++i) {
                uarr[i] = zbinary[i];
            }
            let decompressed = pako.inflate(uarr);
            r = QUtil.uint8array_to_str(decompressed);
        }
        r = eval( "(" + r + ")" );
    }

    if (r.code) {
        let rdataclone = {};
        for(let k in r.data) {
            if (k !== "stack") {
                rdataclone[k] = r.data[k];
            }
        }
        if (r.data.stack) {
            let stack = r.data.stack;
            if (typeof stack !== 'string') {
                stack = JSON.stringify(stack, null, 2);
            }
            msg = r.data.err
                + "\n\n" + JSON.stringify(rdataclone, null, 2)
                + "\n\n@stack:\n" + stack;
        } else {
            msg = JSON.stringify(r.data, null, 2);
        }
        if (this.context.console) {
            this.context.console.text = "[" + r.code + "]\n\n" + msg;
        }
        window.console.error(r);
        throw r;
    } else {
        if (msg && this.context.console) {
            this.context.console.text = msg;
        }
    }
    return r.data;
};

QUtil.prototype.popover = function(selector, msg, position, dismisstime, options) {
    if (!selector) {
        selector = "body";
    }
    let defoptions = {
        html : true,
        content : msg,
        animation : true,
        container : "body",
        placement : position || "auto"
    };

    for (let k in options) {
        defoptions[k] = options[k];
    }

    let target = $(selector);
    target.popover(defoptions);
    target.popover("show");

    if (!dismisstime || dismisstime <= 0) {
        dismisstime = 3000;
    }
    setTimeout(function () {
        $(selector).popover("hide");
    }, dismisstime);
};

QUtil.redis_hash_to_map = function(arr) {
    let m = {};
    for(let i = 0; i < arr.length; i = i + 2) {
        let key = arr[i];
        m[key] = arr[i + 1];
    }
    return m;
};

QUtil.list_merge = function(arr1, arr2) {

    if (!arr1) {
        arr1 = [];
    }
    if (!arr2) {
        arr2 = [];
    }

    let m = {};
    let arr3 = arr1.concat(arr2);
    for(let i = 0; i < arr3.length; i++) {
        let key = arr3[i];
        if (!key) {
            continue;
        }
        m[key] = key;
    }
    let finret = [];
    for(let k in m) {
        finret.push(k);
    }
    return finret;
};

QUtil.get_val = function (o, key, defval) {
    if (!o) {
        return defval;
    }
    let r = o[key];
    if (!r) {
        return defval;
    }
    return r;
};

QUtil.date_add_day = function(d, delta, truncate) {
    let to = d.getTime() + (delta * 24 * 3600 * 1000);
    let r = new Date();
    r.setTime(to);
    if (truncate) {
        r.setHours(0, 0, 0, 0);
    }
    return r;
};



QUtil.date_list_interval = function(from, to, datesplitter, excludes) {

    let ret = [];
    let middle = new Date();
    let one_day = QUtil.DAY_MILLI;
    let from_unix_milli = from.getTime();
    let to_unix_milli = to.getTime();
    middle.setTime(from_unix_milli);
    middle.setHours(23, 59, 59, 999);
    while (true) {
        let include = true;
        if (excludes) {
            let day_of_week = middle.getDay();
            for(let e = 0; e < excludes.length; e++) {
                if (day_of_week === excludes[e]) {
                    include = false;
                    break;
                }
            }
        }
        if (include) {
            let datestr = this.date_format(middle, datesplitter);
            ret.push(datestr);
        }
        middle.setTime(middle.getTime() + one_day);
        if (middle.getTime() >= to_unix_milli) {
            break;
        }
    }
    return ret;
};

QUtil.date_format = function (date, datesplitter) {

    if (typeof datesplitter === 'undefined') {
        datesplitter = "-";
    }
    let y = date.getFullYear();
    let m = date.getMonth() + 1;
    let d = date.getDate();

    let tmp = [];
    tmp.push(y + "");
    tmp.push(datesplitter);
    if (m < 10) {
        tmp.push("0");
    }
    tmp.push(m + "");
    tmp.push(datesplitter);
    if (d < 10) {
        tmp.push("0");
    }
    tmp.push(d + "");

    return tmp.join("");
};


QUtil.time_format = function (date, datesplitter, timesplitter) {
    if (typeof timesplitter === 'undefined') {
        timesplitter = ":";
    }
    let datestr = this.date_format(date, datesplitter);
    let h = date.getHours();
    let m = date.getMinutes();
    let s = date.getSeconds();

    let tmp = [];
    if (h < 10) {
        tmp .push("0");
    }
    tmp.push(h + "");
    tmp.push(timesplitter);
    if (m < 10) {
        tmp.push("0");
    }
    tmp.push(m + "");
    tmp.push(timesplitter);
    if (s < 10) {
        tmp.push("0");
    }
    tmp.push(s + "");
    return datestr + " " + tmp.join("");
};

QUtil.keys = function (m, filter) {
    let keys = [];
    for (let k in m) {
        if (filter) {
            let val = m[k];
            if (filter(m, k, val)) {
                keys.push(k);
            }
        } else {
            keys.push(k);
        }
    }
    return keys;
};

QUtil.values = function (m, filter) {
    let vals = [];
    for (let k in m) {
        let val = m[k];
        if (filter) {
            if (filter(m, k, val)) {
                vals.push(val);
            }
        } else {
            vals.push(val);
        }
    }
    return vals;
};

QUtil.stock_color = function (n) {
    n = n * 1;
    if (isNaN(n) || n === 0) {
        return QUtil.COLOR_EVEN;
    } else if (n > 0) {
        return QUtil.COLOR_UP;
    } else {
        return QUtil.COLOR_DOWN;
    }
};

QUtil.array_field = function(arr, fieldname, discard_null) {
    let ret = [];
    let len = arr.length;
    for(let i = 0; i < len; i++) {
        let one = arr[i];
        let val = one[fieldname];
        if (discard_null && !val) {
            continue;
        }
        ret.push(val);
    }
    return ret;
}

QUtil.array_most = function (arr, callback) {
    let most = arr[0];
    for(let i = 1; i < arr.length; i++) {
        let one = arr[i];
        if (callback(most, one)) {
            most = one;
        }
    }
    return most;
};

QUtil.map_clone = function(m, opt) {
    let clone = {};
    opt = opt || {};
    let ignore = opt.ignore || false;
    for(let k in m) {
        if (ignore && ignore[k]) {
            continue;
        }
        let v = m[k];
        let type = typeof v;
        if (type === 'object' && !opt.obj) {
            continue;
        }
        if (type === 'function' && !opt.func) {
            continue;
        }
        clone[k] = v;
    }
    return clone;
};

QUtil.map_merge = function(des, src, override) {
    for(let k in src) {
        let original = des[k];
        if (original && !override) {
            continue;
        }
        des[k] = src[k];
    }
};

QUtil.map_is_same_by_field_val = function(m1, m2, fields, nrange) {

    if (typeof nrange === 'undefined') {
        nrange = 0.01;
    }

    let same = true;
    for(let i = 0; i < fields.length; i++) {
        let field = fields[i];
        let v1 = m1[field];
        let v2 = m2[field];
        if (!v1 && v2) {
            same = false;
            break;
        }
        let v1n = v1 * 1;
        let v2n = v2 * 1;
        let v1nan = isNaN(v1n);
        let v2nan = isNaN(v2n);
        if ((v1nan && !v2nan) || (v2nan && v1nan)) {
            same = false;
            break;
        }
        if (v1nan && v2nan) {
            if (v1n !== v2n) {
                same = false;
                break;
            }
        } else {
            let delta = Math.abs(v1n - v2n);
            if (delta > nrange) {
                same = false;
                break;
            }
        }
    }
    return same;
};

QUtil.array_clone = function(arr, filter) {
    let clone = [];
    for(let i = 0, n = arr.length; i < n; i++) {
        let one = arr[i];
        if (filter) {
            if (filter(one)) {
                clone[i] = one;
            }
        } else {
            clone[i] = one;
        }
    }
    return clone;
};

QUtil.array_to_map = function (arr, keyname) {
    let m = {};
    let len = arr.length;
    for(let i = 0; i < len; i++) {
        let one = arr[i];
        if (one) {
            let key = one[keyname];
            if (key) {
              m[key] = one;
            }
        }
    }
    return m;
};

QUtil.str_to_uint8array = function(str){
    let raw = str;
    // let raw = window.atob(str);
    let rawlen = raw.length;
    let arr = new Uint8Array(new ArrayBuffer(rawlen));
    for (let i = 0, n = rawlen; i < n; ++i) {
        arr[i] = raw.charCodeAt(i);
    }
    return new Uint8Array(arr);
};


QUtil.uint8array_to_str = function(data){
    let sarr = [];
    for (let i = 0, n = data.length; i < n; i++) {
        sarr.push(String.fromCharCode(data[i]));
    }
    return sarr.join("");
};


QUtil.tree_clone = function(tree, opts) {
    opts =  opts || {};
    let len = tree.length;
    let field_children = opts.field_children || "children";
    opts.current = opts.current || [];
    for (let i = 0; i < len; i++) {
        let clone;
        let one = tree[i];
        if (!one) {
            continue;
        }
        if (opts.cloner) {
            clone = opts.cloner(tree, one, opts);
        } else {
            clone = QUtil.map_clone(one);
        }
        if (clone) {
            opts.current.push(clone);
        } else {
            continue;
        }
        let subtree = one[field_children];
        if (subtree) {
            clone[field_children] = [];
            if (subtree.length > 0) {
                let current = opts.current;
                opts.current = clone[field_children];
                    QUtil.tree_clone(subtree, opts);
                opts.current = current;
            }
        }
    }
    return opts.current;
};

QUtil.tree_locate = function(tree, node, opts) {
    opts = opts || {};
    let len = tree.length;
    let field_id = opts.field_id || "id";
    let field_children = opts.field_children || "children";
    opts.depth = opts.depth || 0;
    if (typeof opts.depth_limit === "undefined") {
        opts.depth_limit = 16;
    }
    opts.pathes = opts.pathes || [];
    for (let i = 0; i < len; i++) {
        let one = tree[i];
        if (one[field_id] === node[field_id]) {
            opts.target = one;
            return opts;
        }

        if (opts.depth + 1 > opts.depth_limit ) {
            continue;
        }

        let subtree = one[field_children];
        if (subtree && subtree.length > 0) {
            opts.pathes[opts.depth] = one;
            opts.depth = opts.depth + 1;
            let ret = QUtil.tree_locate(subtree, node, opts);
            if (ret) {
                return opts;
            }
            opts.pathes[opts.depth] = null;
            opts.depth = opts.depth - 1;
        }
    }
    return null;
};