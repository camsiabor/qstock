

const util = new QUtil();
const vue = new Vue({
    el: '#dcontainer',
    /* [data] ------------------------------------------------------------------- */
    data: {
        db: null,
        stocks: [],
        columns: [],
        setting: {
            mode: "query",
            page_size: 5,
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
        token : {
            snapshot_sz : "meta.a.snapshot.sz",
            snapshot_sh : "meta.a.snapshot.sh",
            khistory_sz : "meta.a.khistory.sz",
            khistory_sh : "meta.a.khistory.sh"
        }
    },
    /* [methods] ------------------------------------------------------------------- */
    methods: {
        config_persist: function () {
            let toptions = $("#table").bootstrapTable("getOptions");
            this.setting.page_size = toptions.pageSize;
            this.setting.script.last = this.script.name;
            if (this.portfolio.name) {
                this.setting.portfolio.last = this.portfolio.name;
                this.setting.portfolio_last = this.portfolio.name;
            }
            let o = {
                columns: this.columns,
                setting: this.setting
            };
            let jstr = JSON.stringify(o, null, 0);
            localStorage.setItem("q.html", jstr);
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

        sync_meta_query: function (dbs, callback) {
            axios.post("/stock/keys", {
                "dbs": dbs,
                "keys" : "meta*"
            }).then(function (resp) {
                let meta = util.handle_response(resp);
                this.db.query_by_id("meta", "id", ["0"], function (data) {
                    let need_refresh = {
                        meta: meta,
                        snapshot_sz: true,
                        snapshot_sh: true,
                        khistory_sz: false,
                        khistory_sh: false
                    };
                    // let current = new Date().getTime() / 1000;
                    let meta_prev = data[0];
                    if (meta_prev) {
                        try {
                            let snapshot_sz = meta[this.token.snapshot_sz]["last"] * 1;
                            let snapshot_sz_prev = meta_prev[this.token.snapshot_sz]["last"] * 1;
                            need_refresh.snapshot_sz = snapshot_sz_prev < snapshot_sz;
                        } catch (e) {
                            need_refresh.snapshot_error = e;
                        }
                        try {
                            let snapshot_sh = meta[this.token.snapshot_sh]["last"] * 1;
                            let snapshot_sh_prev = meta_prev[this.token.snapshot_sh]["last"] * 1;
                            need_refresh.snapshot_sh = snapshot_sh_prev < snapshot_sh;
                        } catch (e) {
                            need_refresh.snapshot_error = e;
                        }
                        try {
                            let khistory_sz = meta[this.token.khistory_sz]["last"] * 1;
                            let khistory_sz_prev = meta_prev[this.token.khistory_sz]["last"] * 1;
                            need_refresh.khistory_sz = khistory_sz_prev < khistory_sz;
                        } catch (e) {
                            need_refresh.khistory_error = e;
                        }
                        try {
                            let khistory_sh = meta[this.token.khistory_sh]["last"] * 1;
                            let khistory_sh_prev = meta_prev[this.token.khistory_sh]["last"] * 1;
                            need_refresh.khistory_sh = khistory_sh_prev < khistory_sh;
                        } catch (e) {
                            need_refresh.khistory_error = e;
                        }

                        if (need_refresh.snapshot_sh || need_refresh.snapshot_sz) {
                            this.db.delete_all("snapshot");
                        }
                    }
                    meta.id = "0";
                    this.db.update("meta", "id", [], meta);
                    if (callback) {
                        callback(need_refresh, meta_prev);
                    }
                }.bind(this));
            }.bind(this)).catch(util.handle_error.bind(this))
        },

        script_list: function () {
            axios.post("/script/list").then(function (json) {
                this.script_names = util.handle_response(json, this.console, "");
            }.bind(this)).catch(util.handle_error.bind(this))
        },

        script_select: function (name) {
            this.script.name = name;
            this.config_persist();
            axios.post("/script/get", {
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
            this.script.script = this.editor.getValue().trim();

            this.config_persist();
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

        stock_sync : function(profile) {
            if (confirm("going to sync? " + profile)) {
                axios.post("/stock/sync", {
                    profile: profile
                }).then(util.handle_response)
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

        stock_get_data_by_code: function (resp) {
            let codes = util.handle_response(resp);
            let has_sz = false;
            let has_sh = false;
            for(let i = 0; i < codes.length; i++) {
                if (codes[i].charAt(0) === '0') {
                    has_sz = true;
                    break;
                }
            }
            for(let i = 0; i < codes.length; i++) {
                if (codes[i].charAt(0) === '6') {
                    has_sh = true;
                    break;
                }
            }
            this.sync_meta_query(["def", "history"], function (need_refresh) {
                let may_i_refresh =
                    (has_sz && need_refresh.snapshot_sz) || (has_sh && need_refresh.snapshot_sh);
                if (may_i_refresh) {
                    axios.post("/stock/gets", {
                        "codes": codes
                    }).then(this.stock_data_adapt);
                } else {
                    let stock_exist = {};
                    for (let i = 0; i < codes.length; i++) {
                        let code = codes[i];
                        stock_exist[code] = false;
                    }
                    this.db.query_by_id("snapshot", "code", codes, function (stocks_local) {
                        for (let i = 0; i < stocks_local.length; i++) {
                            let code = stocks_local[i]["code"];
                            stock_exist[code] = true;
                        }
                        let stock_non_exit = [];
                        for (let code in stock_exist) {
                            if (!stock_exist[code]) {
                                stock_non_exit.push(code);
                            }
                        }
                        if (stock_non_exit.length > 0) {
                            axios.post("/stock/gets", {
                                "codes": codes
                            }).then(function (resp) {
                                let stocks_remote = util.handle_response(resp);
                                resp.data.data = stocks_remote.concat(stocks_local);
                                this.stock_data_adapt(resp);
                            }.bind(this));
                        } else {
                            let resp = {data: {data: stocks_local}};
                            /* fake */
                            this.stock_data_adapt(resp);
                        }
                    }.bind(this));
                }
            }.bind(this));
        },

        stock_data_adapt: function (resp) {
            let stocks = util.handle_response(resp);
            if (stocks instanceof Array) {
                let adata = [];
                let khistory = [];
                for (let i = 0; i < stocks.length; i++) {
                    let stock = stocks[i];

                    stock["turnover"] = util.get_val(stock, "turnover", "").replace("%", "");
                    stock["appointRate"] = util.get_val(stock, "appointRate", "").replace("%", "");
                    adata.push(stock);

                    if (stock.khistory && stock.khistory instanceof Array) {
                        khistory = khistory.concat(stock.khistory);
                    }
                }
                this.init_table(adata);
                this.db.update("snapshot", "code", [], stocks);
                // TODO khistory cache
                // this.db.update("khistory", "code", [ "time" ], khistory);
            } else {
                this.console.text = JSON.stringify(stocks);
            }
        },
        script_query: function () {
            this.console.text = "";
            let script = this.editor.getValue().trim();
            axios.post("/cmd/query", {
                script: script
            }).then(this.stock_get_data_by_code)
        },

        script_test: function () {
            let script = this.editor.getValue().trim();
            axios.post("/cmd/query", {
                script: script
            }).then(function (resp) {
                let data = util.handle_response(resp);
                if (typeof data === 'object') {
                    data = JSON.stringify(data, null, 2);
                }
                this.console.text = data;
            }.bind(this))
        },

        table_resize: function () {
            setTimeout(function () {
                return;
                let browserHeight = document.body.clientHeight;
                let offsetTop = document.getElementById("div_table").offsetTop;
                let theight = browserHeight - offsetTop + 10;
                if (theight < 100) {
                    theight = 100;
                }
                console.log(browserHeight, offsetTop, theight);
                $("#table").bootstrapTable("refreshOptions", {
                    height: theight
                });
            }, 1280);
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
            this.config_persist();
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
        column_setting_show: function () {
            $('#div_column_setting').modal('toggle')
        },
        column_setting_do: function () {
            this.init_table();
            this.config_persist();
            $('#div_column_setting').modal('hide')
        },
        /* [portfolio] ------------------------------------------------------------------- */
        table_get_selection: function (field) {
            let result = [];
            let rows = $("#table").bootstrapTable("getSelections");
            for (let i = 0; i < rows.length; i++) {
                let row = rows[i];
                result.push(row[field]);
            }
            return result;
        },
        portfolio_list: function () {
            axios.post("/cmd/go", {
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
        portfolio_get: function (name, callback) {
            let portfolio_name = "portf_" + name;
            axios.post("/cmd/go", {
                "type": "db",
                "cmd": "Get",
                "args": ["common", "", portfolio_name, true]
            }).then(function (resp) {
                let data = util.handle_response(resp);
                callback.call(this, data);
            }.bind(this));
        },
        portfolio_update: function (codes_sel) {

            if (!this.portfolio.name) {
                util.popover("#button_portfolio_add", "需要组合名字", "bottom");
                return;
            }

            if (!codes_sel) {
                codes_sel = this.table_get_selection("code");
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
            this.config_persist();
            if (this.setting.mode === "portfolio") {
                this.portfolio_view();
            }
        },
        portfolio_unadd: function () {
            let codes = this.table_get_selection("code");
            let portfolio = "portf_" + this.portfolio.name;
            axios.post("/cmd/go", {
                "type": "db",
                "cmd": "Deletes",
                "args": ["common", portfolio, codes]
            }).then(function (resp) {
                util.handle_response(resp);
                // util.popover("body", "删除组合 " + name + " 成功");
            }.bind(this));
        },
        portfolio_view: function (name) {
            if (!name) {
                name = this.portfolio.name;
            }
            this.portfolio_get(name, function (codesm) {
                let codes = util.keys(codesm, function (m, k, v) {
                    return v;
                });
                axios.post("/stock/gets", {
                    "codes": codes
                }).then(this.stock_data_adapt);
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
            axios.post("/cmd/go", {
                "type": "db",
                "cmd": "Delete",
                "args": ["common", "", key]
            }).then(function (resp) {
                util.handle_response(resp);
                this.portfolio.name = "";
                this.portfolio_list();
                util.popover("body", "删除组合 " + name + " 成功");
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
                    this.db.delete_all(target, function (tx) {
                        this.console.text = "web db table " + target + " clear ";
                    }.bind(this));
                    break;
            }
        },


        /* [init] ------------------------------------------------------------------- */

        init_editor: function () {
            this.editor = ace.edit("editor", {
                mode: "ace/mode/lua",
                selectionStyle: "text",
                highlightActiveLine: true,
                highlightSelectedWord: true,
                cursorStyle: "ace",
                newLineMode: "unix",
                fontSize: "0.8em"
            });
            this.editor.setOption("wrap", "free");
            this.editor.setTheme("ace/theme/github");
        },
        init_table: function (data) {

            let columns_show = [];
            for (let i = 0; i < this.columns.length; i++) {
                let col = this.columns[i];
                let coldef = _columns_default_map[col.field];
                if (typeof col.visible === 'undefined') {
                    col.visible = false;
                }
                if (col.field === "operate") {
                    col.visible = true;
                }
                col.sorter = coldef.sorter;
                col.cellStyle = coldef.cellStyle;
                col.formatter = coldef.formatter;
                columns_show.push(col);
            }


            if (this.bstable) {
                if (data) {
                    $("#table").bootstrapTable("refreshOptions", {
                        data: data,
                        columns: columns_show
                    });
                } else {
                    $("#table").bootstrapTable("refreshOptions", {
                        columns: columns_show
                    });
                }
            } else {

                data = data || [
                    {
                        code: "greetings",
                        name: "redis-lua script",
                        nowPrice: 1, openPrice: 1, todayMin: 1, todayMax: 1, swing: 1,
                        diff_money: 1, diff_rate: 1

                    }
                ];

                this.bstable = $("#table").bootstrapTable({
                    data: data,
                    // showToggle :true,
                    // showColumns :true,
                    pagination: true,
                    sidePagination: "client",

                    pageSize: this.setting.page_size || 3,
                    pageList: [3, 5, 7, 10, 25, 50],
                    columns: columns_show,
                    // detailView : true,
                    // detailFormatter : function(index, row, element) {
                    //     return "<span>helloworld</span>";
                    // },
                    onPageChange: function (number, size) {
                        this.setting.page_size = size;
                        this.config_persist();
                    }.bind(this),
                    onDblClickRow: function (row, $element, field, index) {
                        console.log(row, $element, field, index);
                    }.bind(this)
                });

            }

            this.table_resize();
        }
    }, /* methods end */

    /* [mount] ------------------------------------------------------------------- */
    mounted: function () {

        util.context = this;
        // web database
        this.db = new DB({
            "name": "stock"
        });
        this.db.createtable("meta", "id");
        this.db.createtable("snapshot", "code");
        this.db.createtable("khistory", "code", ["date", "data"]);

        // local storage configuration
        this.config_load();

        // view init
        this.init_editor();
        this.init_table();

        this.script_list();
        this.portfolio_list();

        // data init
        this.sync_meta_query(["def", "history"], function () {
            if (this.setting.script.last) {
                this.script_select(this.setting.script.last)
            }
            if (this.setting.portfolio.last || this.setting.portfolio_last) {
                this.portfolio_select(this.setting.portfolio.last || this.setting.portfolio_last);
            }
        }.bind(this));

    }
});
