



Vue.component('vuetable', Vuetable.Vuetable);
Vue.component('vuetable-pagination', Vuetable.VuetablePagination);
Vue.component('treeselect', VueTreeselect.Treeselect);

// Vue.component('vuetable-pagination', Vuetable.VuetablePaginationMixin)
// Vue.component('vuetable-pagination', Vuetable.VuetablePaginationInfoMixin);
// Vue.component('vuetable-pagination-dropdown', Vuetable.VuetablePaginationDropDown);

const def_setting = {
    compress : true,
    table : {
        page_size : 5,
        fields : _columns_default
    },
    indice : {
        num : true,
        kagi : true
    },
    kagi : {
        count : 50,
        width : -2,
        height : 100,
        scale_y : 5,
        lower_field: "exrate",
        lower_scale: 0.33,
        lower_offset: 1
    },
    mode: "query",
    exclude: "buy,sell",
    script: {
        last: "",
        timeout : 300000
    },
    params : {
        last: ""
    },
    portfolio_last: "",
    portfolio: {
        last: ""
    },
    editor : {
        font_size : 1,
        height_mode : "min",
        height_min : 120,
        height_max : 360
    },
    display: {
        editor: true,
        script: true,
        params: true,
        portfolio: true
    }
};

const vue_options = {
    el: '#dcontainer'
};
/* [data] ------------------------------------------------------------------- */
vue_options.data = {
    db: null,
    stocks: [],
    stock : {
        market : "",
        date : "",
        date_offset : 0
    },
    indice : {
        sz : {
            close : 0,
            change_rate : ""
        },
        sh : {
            close : 0,
            change_rate : ""
        }
    },
    calendar : {
        map : { },
        array : [ ]
    },
    columns: [],
    hash : {

    },
    setting: def_setting,
    script_names: [],
    script_setting_opts : {
        type : "",
        act : "",
    },
    script: {
        name: null,
        script: "--[[lua]]--"
    },
    script_group : {
        id : "system",
        name : "system",
        tree : []
    },
    table: {
        data: datamock,
        datamap : {}
    },
    params_names: [],
    params_map : {},
    params: {
        name: "",
        list : [
            {
                "key" : "key",
                "alias" : "别名",
                "value" : "val",
                "expression" : ""
            }
        ]
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
    token: { },
    css : cssmock,

    test : [ {
        id: 'a',
        label: 'a',
        children: [ {
            id: 'aa',
            label: 'aa',
        }, {
            id: 'ab',
            label: 'ab',
        } ],
    }, {
        id: 'b',
        label: 'b',
    }, {
        id: 'c',
        label: 'c',
    } ]

};

/* [computed] --------------------------------------------------------------- */
vue_options.computed = {
    table_fields : function() {
        return this.setting.table.fields;
    }
};

/* [watch] ------------------------------------------------------------------- */
vue_options.watch = {
    "stock.date_offset" : function (n, o) {
        if (n < this.calendar.array.length) {
            this.stock.date = this.calendar.array[n];
        }
        if (!this.stock.date) {
            this.stock.date = QUtil.date_format(new Date(), "");
        }
    },
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
        for (let k in def_setting) {
            if (!this.setting[k]) {
                this.setting[k] = def_setting[k];
            }
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
            switch (target) {
                case "hit":
                case "anti":
                case "debug":
                    this.script_query(target);
                    break;
                case "portfolio":
                    this.portfolio_select();
                    break;
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
        this.table_paging();
        this.editor_init();
        this.stock_indice_fetch();
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
                    window.location.href = window.location.href + "#";
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
        if (!this.editor) {
            this.editor = ace.edit( "editor", {
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
        }
        let height;
        if (this.setting.editor.height_mode === "min") {
            height = this.setting.editor.height_min;
        } else {
            height = this.setting.editor.height_max;
        }
        jQuery("#editor").css("height", height + "px");
        this.editor.setFontSize(this.setting.editor.font_size + "em");
        this.editor.resize()
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


        const fetch_calendar = function() {
            this.stock_calendar_get().then(function () {
                if (this.timer_index_fetch) {
                    clearInterval(this.timer_index_fetch);
                }
                this.timer_index_fetch = setInterval(function () {
                    let date  = new Date();
                    let datestr = QUtil.date_format(date, "");
                    let hours = date.getHours();
                    let minutes = date.getMinutes();
                    if ((hours >= 9 && minutes >= 15) && (hours <= 15)) {
                        if (this.calendar.map[datestr]) {
                            this.stock_indice_fetch();
                        }
                    }
                }.bind(this), 45 * 1000);
            }.bind(this));
        }.bind(this);

        fetch_calendar();

        setInterval(fetch_calendar, 12 * 60 * 60 * 1000);

        this.params_list();
        this.script_list("group").then(function() {
            this.script_list("script").then(function() {
                if (this.setting.script.last) {
                    this.script_select({ id : this.setting.script.last } );
                }
            }.bind(this));
        }.bind(this));
        this.portfolio_list();

        // data init
        this.sync_meta_query(["def", "history"]).then(function () {
            if (this.setting.params.last) {
                this.params_select(this.setting.params.last);
            }
            if (this.setting.portfolio.last || this.setting.portfolio_last) {
                this.portfolio_select(this.setting.portfolio.last || this.setting.portfolio_last);
            }
            this.stock_indice_fetch();
            this.ready = true;
        }.bind(this));
    };

    window.vue = new Vue(vue_options);

}).catch(function (err) {
    console.error("[db]", "[init]", err);
});

