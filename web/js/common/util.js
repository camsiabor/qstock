


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

    // if (!printer) {
    //     if (r.code) {
    //         if (r.data && r.data.stack) {
    //             let rdataclone = {};
    //             for(let k in r.data) {
    //                 if (k !== "stack") {
    //                     rdataclone[k] = r.data[k];
    //                 }
    //             }
    //             window.console.error(rdataclone);
    //             window.console.error(r.data.stack);
    //             throw r.data
    //         }
    //         window.console.error(r);
    //         throw r;
    //     }
    //     return r.data;
    // }

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
        // if (msg === "") {
        //     return r.data;
        // }
        // if (typeof r.data === "object") {
        //     msg = JSON.stringify(r.data, null, 2);
        // } else {
        //     msg = r.data;
        // }
        // if (typeof msg === 'undefined') {
        //     msg = "";
        // }
        // this.context.console.text = msg;
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
        let val = arr[i + 1];
        m[key] = val;
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

QUtil.prototype.format_date = function (date, datesplitter) {

    if (typeof datesplitter === 'undefined') {
        datesplitter = "-";
    }
    let y = date.getFullYear();
    let m = date.getMonth();
    let d = date.getDay();

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


QUtil.prototype.format_time = function (date, datesplitter, timesplitter) {
    if (typeof timesplitter === 'undefined') {
        timesplitter = ":";
    }
    let datestr = this.format_date(date, datesplitter);
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
}

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
}

