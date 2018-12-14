



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
            if (n.table.page_size >= 100) { n.table.page_size = 100; }
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
            if (meta) {
                meta.id = "0";
                this.db.update("meta", [], meta).then();
            }
            return meta;
        }.bind(this))
        .catch(util.handle_error.bind(this));
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



    /* [clear] ------------------------------------------------------------------- */
    clear_cache: function (type, target) {
        switch (type) {
            case "js":
                axios.post("/cmd/go", {
                    "type" : "time",
                    "cmd" : "set",
                    "key" : "js"
                }).then(function () {
                    window.location.href = window.location.href + "";
                });
                break;
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

QUtil.map_merge(vue_options.methods, stock_methods);
QUtil.map_merge(vue_options.methods, script_methods);
QUtil.map_merge(vue_options.methods, portfoilio_methods);

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
            "fields" : [ "_u", "_u_khistory", "khistory_from", "khistory_to", "data" ]
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

