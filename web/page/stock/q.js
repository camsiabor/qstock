



Vue.component('vuetable', Vuetable.Vuetable);
Vue.component('vuetable-pagination', Vuetable.VuetablePagination);
// Vue.component('vuetable-pagination', Vuetable.VuetablePaginationMixin)
// Vue.component('vuetable-pagination', Vuetable.VuetablePaginationInfoMixin);
// Vue.component('vuetable-pagination-dropdown', Vuetable.VuetablePaginationDropDown);



const vue_options = {
    el: '#dcontainer'
};
/* [data] ------------------------------------------------------------------- */
vue_options.data = {
    db: null,
    stocks: [],
    columns: [],
    table: {
        data: datamock,
        datamap : {}
    },
    setting: {
        table : {
            page_size : 5,
            fields : _columns_default
        },
        kagi : {
            count : 20,
            width : -1,
            height : 100,
            scale_y : 1
        },
        mode: "query",
        exclude: "buy,sell",
        script: {
            last: ""
        },
        portfolio_last: "",
        portfolio: {
            last: ""
        },
        display: {
            editor: true,
            script: true,
            portfolio: true
        }
    },
    script_names: [],
    script: {
        name: "",
        script: "--lua.redis"
    },
    portfolio_names: [],
    portfolios: {},
    portfolio: {
        name: ""
    },
    console: {
        text: "redis.lua console",
        msg_error: "error",
        msg_success: "success",
    },
    token: {
        snapshot_sz: "meta.a.snapshot.sz",
        snapshot_sh: "meta.a.snapshot.sh",
        khistory_sz: "meta.a.khistory.sz",
        khistory_sh: "meta.a.khistory.sh"
    },
    css : cssmock

};

/* [computed] --------------------------------------------------------------- */
vue_options.computed = {
    table_fields : function() {
        return this.setting.table.fields;
    }
};

/* [watch] ------------------------------------------------------------------- */
vue_options.watch = {
    "table.data" : function(n, o) {
        for(let i = 0; i < n.length; i++) {
            let one = n[i];
            this.table.datamap[one.code] = one;
        }
        this.$refs.vuetable.refresh();
    },
    setting : {
        handler(n, o) {
            if (!n.table.page_size) { n.table.page_size = 5; }
            if (isNaN(n.table.page_size * 1)) { n.table.page_size = 5; }
            if (n.table.page_size < 0) { n.table.page_size = -n.table.page_size; }
            if (n.table.page_size >= 50) { n.table.page_size = 50; }
            this.config_persist();
        },
        deep: true
    }
};

/* [methods] ------------------------------------------------------------------- */
vue_options.methods = {
    config_persist: function () {

        if (this.timer_config_persist) {
            clearTimeout(this.timer_config_persist);
        }

        this.timer_config_persist = setTimeout(function () {
            let o = {
                columns: this.columns,
                setting: this.setting
            };
            let jstr = JSON.stringify(o, null, 0);
            localStorage.setItem("q.html", jstr);
        }.bind(this), 1000);

    },
    config_load: function () {
        let jstr = localStorage.getItem("q.html");
        if (jstr) {
            let o = JSON.parse(jstr);
            this.columns = o.columns;
            if (o.setting) {
                for (let k in o.setting) {
                    this.setting[k] = o.setting[k];
                }
            }
        } else {
            this.columns = _columns_default;
        }
    },

    sync_meta_query: function (dbs) {
        let meta;
        return axios.post("/stock/keys", {
            "dbs": dbs,
            "keys" : "meta*"
        })
            .then(function (resp) {
                meta = util.handle_response(resp);
                this.db.update("meta",  [], meta).then();
                return meta;
            }.bind(this))
            .catch(util.handle_error.bind(this));
    },

    script_list: function () {
        return axios.post("/script/list").then(function (json) {
            this.script_names = util.handle_response(json, this.console, "");
        }.bind(this)).catch(util.handle_error.bind(this))
    },

    script_select: function (name) {
        this.script.name = name;
        this.setting.script.last = name;
        return axios.post("/script/get", {
            name: name
        }).then(function (resp) {
            let info = util.handle_response(resp);
            this.script.name = info.name;
            this.script.script = info.script;
            this.editor.setValue(this.script.script);
            this.editor.clearSelection();
            if (this.setting.mode === "query") {
                this.script_query();
            }
        }.bind(this)).catch(util.handle_error.bind(this))
    },

    script_save: function () {
        if (!this.script.name) {
            util.popover("#button_script_save", "需要脚本名字", "bottom");
            return;
        }
        this.setting.script.last = this.script.name;
        this.script.script = this.editor.getValue().trim();

        axios.post("/script/update", this.script).then(function (resp) {
            util.handle_response(resp, this.console, "script saved @ " + this.script.name)
            util.popover("#button_script_save", "保存成功", "bottom");
            this.script_list();
        }.bind(this)).catch(util.handle_error.bind(this));
    },

    script_delete: function () {
        if (!confirm("sure to delete? " + this.script.name)) {
            return;
        }
        axios.post("/script/delete", {
            name: this.script.name
        }).then(function (resp) {
            util.handle_response(resp, this.console, "script deleted @ " + this.script.name)
            this.script.name = "---";
            this.script_list();
        }.bind(this)).catch(util.handle_error.bind(this));
    },

    stock_sync : function() {
        let profiles = arguments;
        for(let i = 0; i < profiles.length; i++) {
            let profile = profiles[i];
            if (confirm("going to sync? " + profile)) {
                axios.post("/stock/sync", {
                    profile: profile
                }).then(util.handle_response)
            }
        }
    },

    stock_clear : function(dao, db, group) {
        if (confirm("everything in the database will be clear. sure? " + dao + "." + db + "." + group)) {
            axios.post("/stock/clear", {
                db: db,
                dao: dao,
                group: group
            }).then(function(resp) {
                util.handle_response(resp);
            }.bind(this))
        }
    },

    stock_get_data_by_code: function (resp, time_from, time_to, refresh_view) {

        let codes = util.handle_response(resp);
        if (!codes || !(codes instanceof Array)) {
            if (refresh_view) {
                this.table_init([]);
            }
            return;
        }

        let meta;
        let fetch_khistory = time_from && time_from.length;
        if (fetch_khistory) {
            if (!time_to || !time_to.length) {
                let to = new Date();
                time_to = QUtil.date_format(to, "");
            }
        }
        // TODO short time cache
        let stocks_local;
        return this.sync_meta_query([ "def", "history" ]).then(function (meta_resp) {
            meta = meta_resp;
            return this.db.query_by_id("snapshot",  codes );
        }.bind(this)).then(function(stocks_local_data) {
            stocks_local = stocks_local_data;
            // TODO khistory cache determine
            if (fetch_khistory) {
                let qs = this.db.args_flatten_qs(codes);
                let sql = "SELECT * from khistory where code in (" + qs+ ") AND date >= ? AND date <= ?";
                let args = codes.concat([ time_from, time_to ]);
                let promise = this.db.query(sql, args);
                return promise;
            }
        }.bind(this)).then(function (khistorys) {
            console.log("[meta]", meta);

            let meta_snapshot_last_id = QUtil.get(meta, [ "meta.a.snapshot.sz", "last_id"] , 0) * 1;
            let meta_khistory_last_id_sz = QUtil.get(meta, [ "meta.k.history.sz", "last_id"] , "x").substring(0, 8);
            let meta_khistory_last_id_sh = QUtil.get(meta, [ "meta.k.history.sh", "last_id"] , "x").substring(0, 8);
            let meta_khistory_last_id_ms = QUtil.get(meta, [ "meta.k.history.ms", "last_id"] , "x").substring(0, 8);

            let codes_map = {};
            for (let i = 0; i < codes.length; i++) {
                let code = codes[i];
                codes_map[code] = true;
            }

            // let nowday = new Date().getDay();
            // let is_stock_day = nowday >= 1 && nowday <= 5;
            let stocks_stay = [];
            let codes_fetch = [];
            for (let i = 0; i < stocks_local.length; i++) {
                let stock = stocks_local[i];
                let code = stock["code"];
                let _u = stock["_u"] * 1;
                let stay = false;
                if (meta_snapshot_last_id === _u) {
                    if (fetch_khistory) {
                        let _u_khistory = stock["_u_khistory"];
                        if (_u_khistory) {
                            _u_khistory = _u_khistory.substring(0, 8);
                        }
                        switch (code.charAt(0)) {
                            case '0':
                                stay = (_u_khistory === meta_khistory_last_id_sz);
                                break;
                            case '6':
                                stay = (_u_khistory === meta_khistory_last_id_sh);
                                break;
                            default:
                                stay = (_u_khistory === meta_khistory_last_id_ms);
                                break;
                        }
                    } else {
                        stay = true;
                    }
                }
                if (stay) {
                    stocks_stay.push(stock);
                    codes_map[code] = false;
                }
            }

            if (fetch_khistory && khistorys && khistorys.length && stocks_stay.length) {
                let stocks_stay_map = QUtil.array_to_map(stocks_stay, "code");
                for(let i = 0; i < khistorys.length; i++) {
                    let one = khistorys[i];
                    let code = one.code;
                    let stock = stocks_stay_map[code];
                    if (stock) {
                        stock.khistory = stock.khistory || [];
                        stock.khistory.push(one);
                    }
                }
            }

            for(let code in codes_map) {
                if (codes_map[code]) {
                    codes_fetch.push(code);
                }
            }
            let wrap = {
                stocks : stocks_stay,
                stocks_local : stocks_stay,
                codes_fetch : codes_fetch,
                time_from : time_from,
                time_to : time_to,
                meta : meta,
                refresh_view : refresh_view
            };
            if (codes_fetch.length > 0) {
                return this.stock_data_request(wrap);
            } else {
                return this.stock_data_adapt(wrap);
            }
        }.bind(this));
    },

    stock_data_request: function(wrap) {

        return axios.post("/stock/gets", {
            "codes": wrap.codes_fetch,
            "time_from" : wrap.time_from,
            "time_to" : wrap.time_to
        }).then(function (resp) {
            wrap.stocks = util.handle_response(resp);
            if (wrap.stocks_local) {
                wrap.stocks = wrap.stocks.concat(wrap.stocks_local);
            }
            return this.stock_data_adapt(wrap);
        }.bind(this));
    },

    stock_data_adapt: function (wrap) {
        let refresh_view = wrap.refresh_view;
        if (typeof refresh_view === 'undefined') {
            refresh_view = true;
        }
        let stocks = wrap.stocks;
        let stocks_local = wrap.stocks_local;
        if (!stocks instanceof Array) {
            this.console.text = JSON.stringify(stocks);
            return;
        }
        // let max_date = "";
        let view_data = [];
        let update_data = [];
        let khistorys = [];
        let khistorys_map = {};
        let stocks_local_map = QUtil.array_to_map(stocks_local, "code");
        let meta = wrap.meta;
        let meta_khistory_last_id_sz = QUtil.get(meta, [ "meta.k.history.sz", "last_id"] , "-");
        let meta_khistory_last_id_sh = QUtil.get(meta, [ "meta.k.history.sh", "last_id"] , "-");
        let meta_khistory_last_id_ms = QUtil.get(meta, [ "meta.k.history.ms", "last_id"] , "-");
        for (let i = 0; i < stocks.length; i++) {
            let stock = stocks[i];
            let code = stock.code;
            let stock_khistory = stock.khistory;
            if (!stocks_local_map[code]) {
                update_data.push(stock);
                if (stock_khistory && stock_khistory.length > 0) {
                    for(let k = 0; k < stock_khistory.length; k++) {
                        let onek = stock_khistory[k];
                        onek.id = code + "-" + onek.date;
                    }
                    khistorys = khistorys.concat(stock_khistory);
                    khistorys_map[code] = stock_khistory;
                    // if (!max_date) {
                    //     let max = util.array_most(stock_khistory, function (max, one) {
                    //         return max.date * 1 > one.date * 1;
                    //     });
                    //     max_date = max.date;
                    // }
                    switch(code.charAt(0)) {
                        case '0':
                            stock._u_khistory = meta_khistory_last_id_sz;
                            break;
                        case '6':
                            stock._u_khistory = meta_khistory_last_id_sh;
                            break;
                        default:
                            stock._u_khistory = meta_khistory_last_id_ms;
                            break;
                    }
                    stock.khistory = null;
                }
            }
            view_data.push(stock);
        }


        let promise;
        if (update_data.length) {
            let fields_update = [ "_u" ];
            if (khistorys.length) {
                fields_update.push("_u_khistory");
            }
            promise = this.db.update("snapshot", fields_update, update_data);
        } else {
            promise = Promise.resolve(stocks);
        }
        return promise.then(function () {
            if (khistorys.length > 0) {
                return this.db.update("khistory", ["code", "date" ], khistorys);
            }
            return stocks;
        }.bind(this)).then(function () {
            return stocks;
        }).then(function () {
            if (khistorys.length > 0) {
                for (let i = 0; i < stocks.length; i++) {
                    let stock = stocks[i];
                    let code = stock.code;
                    stock.khistory = khistorys_map[code];
                }
            }
            if (refresh_view) {
                this.table_init(view_data);
            }
            return stocks;
        }.bind(this));
    },

    script_query: function () {
        this.console.text = "";
        let script = this.editor.getValue().trim();
        return axios.post("/cmd/query", {
            script: script
        }).then(this.stock_get_data_by_code)
    },

    script_test: function () {
        let script = this.editor.getValue().trim();
        return axios.post("/cmd/query", {
            script: script
        }).then(function (resp) {
            let data = util.handle_response(resp);
            if (typeof data === 'object') {
                data = JSON.stringify(data, null, 2);
            }
            this.console.text = data;
            return data;
        }.bind(this))
    },

    toggle: function (target, type) {
        if (type === "view") {
            this.setting.display[target] = !this.setting.display[target];
        } else if (type === "mode") {
            this.setting.mode = target;
            if (target === "query") {
                this.script_select();
            } else if (target === "portfolio") {
                this.portfolio_select();
            } else {

            }
        }
    },
    isexclude: function (s) {

        let arr = this.setting.exclude.split(",")
        for (let i = 0; i < arr.length; i++) {
            let sub = arr[i];
            if (sub) {
                if (s.indexOf(sub) >= 0) {
                    return true;
                }
            }
        }
        return false;
    },
    setting_show: function () {
        $('#div_setting').modal('toggle');
    },
    setting_save: function () {
        $('#div_setting').modal('hide');
        // this.table_init(this.table.data);
        this.table_paging();
    },
    /* [portfolio] ------------------------------------------------------------------- */
    table_get_selection: function (retrow) {
        let codes = [].concat(this.$refs.vuetable.selectedTo);
        if (retrow) {
            let result = [];
            for (let i = 0; i < codes.length; i++) {
                let code = codes[i]
                let one = this.table.datamap[code];
                if (one) {
                    result.push(one);
                }
            }
            return result;
        } else {
            return codes;
        }
    },
    portfolio_list: function () {
        return axios.post("/cmd/go", {
            "type": "db",
            "cmd": "Keys",
            "args": ["common", "", "portf_*"],
        }).then(function (resp) {
            let data = util.handle_response(resp);
            for (let i = 0; i < data.length; i++) {
                let name = data[i];
                name = name.replace("portf_", "");
                data[i] = name;
            }
            this.portfolio_names = data;
        }.bind(this));
    },

    portfolio_update: function (codes_sel) {

        if (!this.portfolio.name) {
            util.popover("#button_portfolio_add", "需要组合名字", "bottom");
            return;
        }

        if (!codes_sel) {
            codes_sel = this.table_get_selection();
            if (codes_sel.length === 0) {
                util.popover("#button_portfolio_add", "需要选择对象", "bottom");
                return;
            }
        }

        let portfolio_name = "portf_" + this.portfolio.name;
        axios.post("/cmd/go", {
            "type": "db",
            "cmd": "Updates",
            "args": ["common", portfolio_name, codes_sel, codes_sel, true, false]
        }).then(function (resp) {
            util.handle_response(resp, this.console, "");
            let msg = "加入到 " + this.portfolio.name + " 成功"
            util.popover("#input_portfolio_name", msg, "bottom")
            this.portfolio_list();
        }.bind(this));
    },
    portfolio_add_manual: function () {
        let codestr = prompt("编号可以用逗号分隔");
        codestr = codestr.trim();
        codestr = codestr.split(",");
        let valid = [];
        for (let i = 0; i < codestr.length; i++) {
            let code = codestr[i].trim();
            if (code.length === 6 && !isNaN(code * 1)) {
                valid.push(code);
            }
        }
        this.portfolio_update(valid);
    },
    portfolio_select: function (pname) {
        if (pname) {
            this.portfolio.name = pname;
        }
        this.setting.portfolio.last = pname;
        if (this.setting.mode === "portfolio") {
            this.portfolio_view();
        }
    },
    portfolio_unadd: function (codes) {
        if (!codes) {
            codes = this.table_get_selection();
        }
        if (codes.length === 0) {
            alert("select something please");
            return;
        }
        let portfolio = "portf_" + this.portfolio.name;
        return axios.post("/cmd/go", {
            "type": "db",
            "cmd": "Deletes",
            "args": ["common", portfolio, codes]
        }).then(function (resp) {
            util.handle_response(resp);
            return this.portfolio_view();
        }.bind(this));
    },
    portfolio_view: function (name) {
        if (!name) {
            name = this.portfolio.name;
        }
        let portfolio_name = "portf_" + name;
        return axios.post("/cmd/go", {
            "type": "db",
            "cmd": "Get",
            "args": ["common", "", portfolio_name, true]
        }).then(function (resp) {
            let data = util.handle_response(resp);
            let codes = QUtil.keys(data, function (m, k, v) {
                return v;
            });
            return this.stock_get_data_by_code(codes);
        }.bind(this));
    },
    portfolio_delete: function (name) {
        if (!name) {
            name = this.portfolio.name;
        }
        if (!confirm("sure to delete portfolio? " + name)) {
            return;
        }
        let key = "portf_" + name;
        return axios.post("/cmd/go", {
            "type": "db",
            "cmd": "Delete",
            "args": ["common", "", key]
        }).then(function (resp) {
            util.handle_response(resp);
            this.portfolio.name = "";
            return this.portfolio_list();
        }.bind(this));
    },


    /* [clear] ------------------------------------------------------------------- */
    clear_cache: function (type, target) {
        switch (type) {
            case "local" :
                window.localStorage.clear();
                this.console.text = "local storage clear";
                break;
            case "webdb":
                this.db.delete_all(target).then(function () {
                    this.console.text = "web db table " + target + " clear ";
                }.bind(this));
                break;
        }
    },



    editor_init: function () {
        this.editor = ace.edit("editor", {
            mode: "ace/mode/lua",
            selectionStyle: "text",
            highlightActiveLine: true,
            highlightSelectedWord: true,                cursorStyle: "ace",
            newLineMode: "unix",
            fontSize: "0.8em"
        });
        this.editor.setOption("wrap", "free");
        this.editor.setTheme("ace/theme/github");
    },

    chart_kagi_init: function (codes) {
        if (!codes) {
            codes = QUtil.array_field(this.table.stocks_view, "code");
        }
        let vuetable = this.$refs.vuetable;
        let children = vuetable.$children;
        let chart_children = [];
        for(let i = 0; i < children.length; i++) {
            let one = children[i];
            let code = one.chart_render && one.rowData && one.rowData.code;
            if (code) {
                // codes.push(code);
                chart_children.push(one);
            }
        }

        if (codes.length > 0) {
            let kagi_setting = this.setting.kagi;

            let now = new Date();
            let from = QUtil.date_add_day(now, -kagi_setting.count);
            let time_from = QUtil.date_format(from, "");
            this.stock_get_data_by_code(codes, time_from, "", false).then(function (stocks) {
                let stocks_map = QUtil.array_to_map(stocks, "code");
                stocks_map.kagi = kagi_setting;
                for(let i = 0; i < chart_children.length; i++) {
                    let one = chart_children[i];
                    try {
                        one.chart_render(stocks_map);
                    } catch (ex) {
                        console.error(ex);
                    }
                }
            });
        }

    },

    table_init: function (data) {

        // TODO columns
        let columns_show = [];
        for (let i = 0; i < this.columns.length; i++) {
            let col = this.columns[i];
            if (typeof col.visible === 'undefined') {
                col.visible = false;
            }
            if (col.field === "operate") {
                col.visible = true;
            }
            let coldef = _columns_default_map[col.field];
            if (!coldef) {
                continue;
            }
            col.sorter = coldef.sorter;
            col.callback = coldef.callback;
            columns_show.push(col);
        }


        if (data) {
            this.table.data = data;
        }
    },

    table_paging(data) {
        data = data || this.table.pagination;
        this.$refs.pagination.setPaginationData(data);
        // let stocks = this.table.data.slice(data.from - 1, data.to);
        // let codes = QUtil.array_field(stocks, "code");
        this.chart_kagi_init();
    },

    table_paging_change(page) {
        this.$refs.vuetable.changePage(page)
    },

    table_data_manage(sortOrder, pagination) {
        let data = this.table.data;
        if (data.length <= 0) {
            return;
        }

        // sortOrder can be empty, so we have to check for that as well
        let dataview = [].concat(data);
        let sortlen = sortOrder.length;
        if (sortlen > 0) {
            /*
            direction, field, sortField
             */

            for(let i = 0; i < sortlen; i++) {
                let one = sortOrder[i];
                one.asc = one.direction === 'asc';
            }

            dataview = dataview.sort(function(a , b){
                let r = 0;
                for(let i = 0; i < sortlen; i++) {
                    let one = sortOrder[i];
                    let field = one.sortField;
                    if (one.asc) {
                        r = a[field] * 1 - b[field] * 1;
                    } else {
                        r = b[field] * 1 - a[field] * 1;
                    }
                    if (r !== 0) {
                        return r;
                    }
                }
            })
        }
        let page_size = this.setting.table.page_size * 1;
        pagination = this.$refs.vuetable.makePagination(
            data.length,
            page_size
        );

        let from = pagination.from - 1;
        let to = from + page_size;
        dataview = dataview.slice(from, to);

        this.table.stocks_view = dataview;
        this.table.pagination = pagination;
        return {
            pagination: pagination,
            data: dataview
        };
    }
}; /* methods end */

/* ==================== web database init ===================== */

DB.new_db_promise({
    name: "stock",
    dbsize : 32 * 1024 * 1024,
    schema : {
        "meta" : {
            "keyname" : "id"
        },
        "snapshot" : {
            "keyname" : "code",
            "fields" : [ "_u", "_u_khistory", "data" ]
        },
        "khistory" : {
            "keyname" : "id",
            "fields" : [ "code", "date", "data" ]
        }
    }
}).then(function (db) {

    vue_options.mounted = function () {

        util.context = this;

        this.db = db;

        // local storage configuration
        this.config_load();

        // view init
        this.editor_init();
        this.table_init();

        this.script_list();
        this.portfolio_list();

        // data init
        this.sync_meta_query(["def", "history"]).then(function () {
            if (this.setting.script.last) {
                this.script_select(this.setting.script.last)
            }
            if (this.setting.portfolio.last || this.setting.portfolio_last) {
                this.portfolio_select(this.setting.portfolio.last || this.setting.portfolio_last);
            }
        }.bind(this));
    };
    window.vue = new Vue(vue_options);

}).catch(function (err) {
    console.error("[db]", "[init]", err);
});

