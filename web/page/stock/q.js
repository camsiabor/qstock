


Vue.component('vuetable-actions', {
    template : 
        "<div class='btn-group btn-group-sm'>" +
            "<button class='btn btn-sm btn-outline-secondary' @click='act(\"view.detail\", rowData, rowIndex)'><i class='fa fa-search s-tiny'></i></button>" +
            "<button class='btn btn-sm btn-outline-secondary' @click='act(\"portfolio.add\", rowData, rowIndex)'><i class='fa fa-plus-circle s-tiny'></i></button>" +
            "<button class='btn btn-sm btn-outline-secondary' @click='act(\"portfolio.unadd\", rowData, rowIndex)'><i class='fa fa-minus-circle s-tiny'></i></button>" +
        "</div>",
    template2 :
        "<div class=''>" +
        "<div><button class='btn btn-sm btn-outline-secondary' @click='act(\"view.detail\", rowData, rowIndex)'><i class='fa fa-search s-tiny'></i></button></div>" +
        "<div><button class='btn btn-sm btn-outline-secondary' @click='act(\"portfolio.add\", rowData, rowIndex)'><i class='fa fa-plus-circle s-tiny'></i></button></div>" +
        "<div><button class='btn btn-sm btn-outline-secondary' @click='act(\"portfolio.unadd\", rowData, rowIndex)'><i class='fa fa-minus-circle s-tiny'></i></button></div>" +
        "</div>",
    props: {
        rowData: {
            type: Object,
            required: true
        },
        rowIndex: {
            type: Number
        }
    },
    methods: {
        act (action, data, index) {
            let code = data.code;
            let context = this.$root;
            switch (action) {
                case "portfolio.add":
                    context.portfolio_update( [ code ]);
                    break;
                case "portfolio.unadd":
                    context.portfolio_unadd( [ code ]);
                    break;
            }
            console.log('custom-actions: ' + action, data.name, index, data);
        }
    }
});

Vue.component('vuetable', Vuetable.Vuetable);
Vue.component('vuetable-pagination', Vuetable.VuetablePagination);
// Vue.component('vuetable-pagination', Vuetable.VuetablePaginationMixin)
// Vue.component('vuetable-pagination', Vuetable.VuetablePaginationInfoMixin);
// Vue.component('vuetable-pagination-dropdown', Vuetable.VuetablePaginationDropDown);

const vue = new Vue({
    el: '#dcontainer',
    /* [data] ------------------------------------------------------------------- */
    data() {
        return {
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
                    count : 20
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
        }
    },
    /* [methods] ------------------------------------------------------------------- */
    methods: {
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
            let profiles_str = "";
            for(let i = 0; i < profiles.length; i++) {
                profiles_str = profiles_str + " " + profiles[i];
            }
            if (confirm("going to sync? " + profiles_str)) {
                for(let i = 0; i < profiles.length; i++) {
                    axios.post("/stock/sync", {
                        profile: profiles[i]
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

        stock_get_data_by_code: function (resp, refresh_table, fetch_khistory, time_from, time_to) {

            if (typeof fetch_khistory === 'undefined') {
                fetch_khistory = false;
            }

            let codes;
            if (resp instanceof Array) {
                codes = resp;
            } else {
                codes = util.handle_response(resp);
            }
            if (!codes || !(codes instanceof Array)) {
                if (refresh_table) {
                    this.table_init([]);
                }
                return;
            }

            let codes_map = {};
            for (let i = 0; i < codes.length; i++) {
                let code = codes[i];
                codes_map[code] = false;
            }

            let meta;
            this.sync_meta_query([ "def", "history" ]).then(function (meta_resp) {
                meta = meta_resp;
                return this.db.query_by_id("snapshot",  codes );
            }.bind(this)).then(function (stocks_local) {
                let nowday = new Date().getDay();
                let is_stock_day = nowday >= 1 && nowday <= 5;

                for (let i = 0; i < stocks_local.length; i++) {
                    let stock = stocks_local[i];
                    let code = stock["code"];
                    if (is_stock_day) {
                        let _u = stock["_u"];
                    } else {
                        codes_map[code] = true;
                    }

                }
                let codes_need_refresh = [];
                for (let code in codes_map) {
                    if (!codes_map[code]) {
                        codes_need_refresh.push(code);
                    }
                }
                if (codes_need_refresh.length > 0) {
                    this.stock_data_request(codes_need_refresh, stocks_local, time_from, time_to)
                } else {
                    let resp = {data: {data: stocks_local}};
                    /* fake */
                    this.stock_data_adapt(resp, false, true);
                }
            }.bind(this));
        },

        stock_data_request: function(codes, stocks, time_from, time_to, fetch_khistory, refresh_table) {

            if (typeof time_to === "string" && time_to.length === 0) {
                let to = new Date();
                time_to = util.date_format(to, "");
            }

            if (typeof time_from === 'string' && time_from.length === 0) {
                let now = new Date();
                let from = util.date_add_day(now, -30);
                time_from = util.date_format(from, "");
            }
            if (typeof fetch_khistory === 'undefined') {
                fetch_khistory = false;
            }
            if (typeof refresh_table === 'undefined') {
                refresh_table = true;
            }

            return axios.post("/stock/gets", {
                "codes": codes,
                "time_from" : time_from,
                "time_to" : time_to
            }).then(function (resp) {
                let stocks_remote = util.handle_response(resp);
                if (stocks) {
                    resp.data.data = stocks_remote.concat(stocks);
                }
                this.stock_data_adapt(resp, fetch_khistory, refresh_table);
            }.bind(this));
        },

        stock_data_adapt: function (resp, fetch_khistory, refresh_table) {
            let stocks = util.handle_response(resp);
            if (!stocks instanceof Array) {
                this.console.text = JSON.stringify(stocks);
                return;
            }
            let adata = [];
            let khistory = [];
            let khistory_map = {};
            let max_date = "";
            for (let i = 0; i < stocks.length; i++) {
                let stock = stocks[i];
                let code = stock.code;
                let stock_khistory = stock.khistory;

                adata.push(stock);
                if (stock_khistory && stock_khistory.length > 0) {
                    for(let k = 0; k < stock_khistory.length; k++) {
                        let onek = stock_khistory[k];
                        onek.id = code + "-" + onek.date;
                    }
                    khistory = khistory.concat(stock_khistory);
                    khistory_map[code] = stock_khistory;
                    if (!max_date) {
                        let max = util.array_most(stock_khistory, function (max, one) {
                            return max.date * 1 > one.date * 1;
                        });
                        max_date = max.date;
                    }
                    stock._u_khistory = max_date;
                    stock.khistory = null;
                }
            }

            return this.db.update("snapshot", [ "_u" ], stocks).then(function () {
                if (khistory.length > 0) {
                    for (let i = 0; i < stocks.length; i++) {
                        let stock = stocks[i];
                        let code = stock.code;
                        stock.khistory = khistory_map[code];
                    }
                }
                if (khistory.length > 0) {
                    return this.db.update("khistory", ["code", "date" ], khistory);
                }
            }.bind(this)).then(function () {
               // khistory update promise
            }).finally(function () {
                if (refresh_table) {
                    this.table_init(adata);
                }
            }.bind(this));

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
        column_setting_show: function () {
            $('#div_column_setting').modal('toggle')
        },
        column_setting_do: function () {
            this.table_init();
            $('#div_column_setting').modal('hide')
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
            axios.post("/cmd/go", {
                "type": "db",
                "cmd": "Deletes",
                "args": ["common", portfolio, codes]
            }).then(function (resp) {
                util.handle_response(resp);
                this.portfolio_view();
                // util.popover("body", "删除组合 " + name + " 成功");
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
                let codes = util.keys(data, function (m, k, v) {
                    return v;
                });
                this.stock_data_request(codes);
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
                    this.db.delete_all(target).then(function () {
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
                highlightSelectedWord: true,                cursorStyle: "ace",
                newLineMode: "unix",
                fontSize: "0.8em"
            });
            this.editor.setOption("wrap", "free");
            this.editor.setTheme("ace/theme/github");
        },

        chart_init: function () {
            let vuetable = this.$refs.vuetable;
            let children = vuetable.$children;
            for(let i = 0; i < children.length; i++) {
                let one = children[i];
                if (typeof one.chart_request === 'function') {
                    one.chart_request();
                }
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
            this.$refs.pagination.setPaginationData(data);
            this.chart_init();
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
            let clone = [].concat(data);
            let sortlen = sortOrder.length;
            if (sortlen > 0) {
                /*
                direction, field, sortField
                 */

                for(let i = 0; i < sortlen; i++) {
                    let one = sortOrder[i];
                    one.asc = one.direction === 'asc';
                }

                clone = clone.sort(function(a , b){
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
            clone = clone.slice(from, to);
            return {
                pagination: pagination,
                data: clone
            };
        },

    }, /* methods end */


    /* [computed] --------------------------------------------------------------- */

    computed : {

        table_fields : function() {
            return this.setting.table.fields;
        },

        // data : function () {
        //     return { data : this.table.data };
        // }

    },

    /* [watch] ------------------------------------------------------------------- */

    watch: {
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
    },

    /* [mount] ------------------------------------------------------------------- */
    mounted () {

        util.context = this;
        // web database
        DB.new_db_promise({
            name: "stock",
            dbsize : 32 * 1024 * 1024,
            schema : {
                "meta" : {
                    "keyname" : "id"
                },
                "snapshot" : {
                    "keyname" : "code",
                    "fields" : [ "_u", "_u_history", "data" ]
                },
                "khistory" : {
                    "keyname" : "id",
                    "fields" : [ "code", "date", "data" ]
                }
            }
        }).then(function (db) {
            this.db = db;

            // local storage configuration
            this.config_load();

            // view init
            this.init_editor();
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

        }.bind(this))
        .catch(function (err) {
           console.error("[fatal]", err);
        });

    } // mount end
});
