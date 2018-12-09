
QUtil.DAY_MILLI = 24 * 3600 * 1000;

function QUtil(opt) {
    opt = opt || {};
    for(let k in opt) {
        this[k] = opt[k];
    }
    this.context = this.context || window;
}

QUtil.prototype.handle_error = function(error) {
    let err = error.stack || error;
    if (this.context.console) {
        this.context.console.text = err;
    }
    window.console.error(err);
    this.popover("", err);
};

QUtil.prototype.handle_response = function(resp, printer, msg) {
    let r = resp.data;
    if (typeof r === "string") {
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
            msg = r.data.err
                + "\n\n" + JSON.stringify(rdataclone, null, 2)
                + "\n\n" + r.data.stack;
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

QUtil.prototype.redis_hash_to_map = function(arr) {
    let m = {};
    for(let i = 0; i < arr.length; i = i + 2) {
        let key = arr[i];
        m[key] = arr[i + 1];
    }
    return m;
};

QUtil.prototype.list_merge = function(arr1, arr2) {

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

QUtil.prototype.get_val = function (o, key, defval) {
    if (!o) {
        return defval;
    }
    let r = o[key];
    if (!r) {
        return defval;
    }
    return r;
};

QUtil.prototype.date_add_day = function(d, delta, truncate) {
    let to = d.getTime() + (delta * 24 * 3600 * 1000);
    let r = new Date();
    r.setTime(to);
    if (truncate) {
        r.setHours(0, 0, 0, 0);
    }
    return r;
};



QUtil.prototype.date_list_interval = function(from, to, datesplitter, excludes) {

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

QUtil.prototype.date_format = function (date, datesplitter) {

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


QUtil.prototype.time_format = function (date, datesplitter, timesplitter) {
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

QUtil.prototype.keys = function (m, filter) {
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

QUtil.prototype.values = function (m, filter) {
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

QUtil.prototype.stock_color = function (n) {
    n = n * 1;
    if (isNaN(n) || n === 0) {
        return "grey";
    } else if (n > 0) {
        return "red";
    } else {
        return "green";
    }
};

QUtil.prototype.array_most = function (arr, callback) {
    let max = arr[0];
    for(let i = 1; i < arr.length; i++) {
        let one = arr[i];
        if (callback(max, one)) {
            max = one;
        }
    }
    return max;
};