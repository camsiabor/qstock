


const stock_methods = {
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
        let do_fetch_khistory = time_from && time_from.length;
        if (do_fetch_khistory) {
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
            if (do_fetch_khistory) {
                let qs = this.db.args_flatten_qs(codes);
                let sql = "SELECT * from khistory where code in (" + qs+ ") AND date >= ? AND date <= ?";
                let args = codes.concat([ time_from, time_to ]);
                let promise = this.db.query(sql, args);
                return promise;
            }
        }.bind(this)).then(function (khistorys) {
            console.log("[meta]", meta);

            let meta_snapshot_last_id = QUtil.get(meta, [ "meta.a.snapshot.sz", "last_id"] , 0) * 1;
            // let meta_khistory_last_id_sz = QUtil.get(meta, [ "meta.k.history.sz", "last_id"] , "x").substring(0, 8);
            // let meta_khistory_last_id_sh = QUtil.get(meta, [ "meta.k.history.sh", "last_id"] , "x").substring(0, 8);
            // let meta_khistory_last_id_su = QUtil.get(meta, [ "meta.k.history.su", "last_id"] , "x").substring(0, 8);

            let codes_map = {};
            for (let i = 0; i < codes.length; i++) {
                let code = codes[i];
                codes_map[code] = {
                    "do" : true,
                    "code" : code
                };
            }

            // let nowday = new Date().getDay();
            // let is_stock_day = nowday >= 1 && nowday <= 5;
            let stocks_stay = [];
            let fetch_pending = [];
            let time_to_num = time_to * 1;
            let time_from_num = time_from * 1;
            for (let i = 0; i < stocks_local.length; i++) {
                let stock = stocks_local[i];
                let code = stock["code"];
                let _u = stock["_u"] * 1;
                let stay = true;
                let fetch_time_from, fetch_time_to;
                if (meta_snapshot_last_id !== _u) {
                    stay = false;
                }
                if (do_fetch_khistory) {
                    let khistory_to = (stock["khistory_to"] || (time_to_num - 1) ) * 1;
                    let khistory_from = (stock["khistory_from"] || (time_from_num + 1)) * 1;
                    if (khistory_from <= time_from_num && khistory_to >= time_to_num) {
                        fetch_time_from = "x";
                    } else {
                        stay = false;
                        fetch_time_to = time_to;
                        fetch_time_from = time_from;
                        if (khistory_to >= time_to_num) {
                            fetch_time_to = khistory_from + "";
                        }
                        if (khistory_from <= time_from_num) {
                            fetch_time_from = khistory_to + "";
                        }
                    }
                }

                let fetch_one = codes_map[code];
                if (stay) {
                    fetch_one.do = false;
                    stocks_stay.push(stock);
                } else {
                    fetch_one.do = true;
                    if (do_fetch_khistory) {
                        if (fetch_time_from) {
                            fetch_one.from = fetch_time_from;
                        }
                        if (fetch_time_to) {
                            fetch_one.to = fetch_time_to;
                        }
                    }
                }
            }

            for(let code in codes_map) {
                let fetch_one = codes_map[code];
                if (fetch_one.do) {
                    delete fetch_one.do;
                    fetch_pending.push(fetch_one);
                }
            }

            let khistory_map ={};
            if (do_fetch_khistory && khistorys && khistorys.length) {
                let stocks_stay_map = QUtil.array_to_map(stocks_stay, "code");
                for(let i = 0; i < khistorys.length; i++) {
                    let one = khistorys[i];
                    let code = one.code;
                    let stock = stocks_stay_map[code];
                    if (stock) {
                        stock.khistory = stock.khistory || [];
                        stock.khistory.push(one);
                        khistory_map[code] = stock.khistory;
                    }
                }
            }

            let wrap = {
                stocks : stocks_stay,
                stocks_local : stocks_stay,
                fetch_pending : fetch_pending,
                time_from : time_from,
                time_to : time_to,
                meta : meta,
                refresh_view : refresh_view,
                khistory_local : khistorys,
                khistory_map : khistory_map
            };
            if (fetch_pending.length > 0) {
                return this.stock_data_request(wrap);
            } else {
                return this.stock_data_adapt(wrap);
            }
        }.bind(this));
    },

    stock_data_request: function(wrap) {
        return axios.post("/stock/gets", {
            "fetchs": wrap.fetch_pending,
            "time_from" : wrap.time_from,
            "time_to" : wrap.time_to
        }).then(function (resp) {
            wrap.stocks = util.handle_response(resp);
            if (wrap.stocks_local && wrap.stocks_local.length) {
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
        let khistory = [];
        let khistory_map = wrap.khistory_map;
        let stocks_local_map = QUtil.array_to_map(stocks_local, "code");
        let meta = wrap.meta;
        let meta_khistory_last_id_sz = QUtil.get(meta, [ "meta.k.history.sz", "last_id"] , "-");
        let meta_khistory_last_id_sh = QUtil.get(meta, [ "meta.k.history.sh", "last_id"] , "-");
        let meta_khistory_last_id_su = QUtil.get(meta, [ "meta.k.history.su", "last_id"] , "-");
        for (let i = 0; i < stocks.length; i++) {
            let stock = stocks[i];
            let code = stock.code;
            if (!code) {
                continue;
            }
            if (!stocks_local_map[code]) {
                update_data.push(stock);
                if (wrap.time_from) {
                    let stock_khistory = stock.khistory || [];
                    for(let k = 0; k < stock_khistory.length; k++) {
                        let onek = stock_khistory[k];
                        onek.id = code + "-" + onek.date;
                    }
                    if (stock_khistory.length) {
                        khistory = khistory.concat(stock_khistory);
                    }
                    let stock_khistory_local = khistory_map[code];
                    if (stock_khistory_local) {
                        stock_khistory = stock_khistory_local.concat(stock_khistory);
                    } else {
                        khistory_map[code] = stock_khistory;
                    }
                    switch(code.charAt(0)) {
                        case '0':
                            stock._u_khistory = meta_khistory_last_id_sz;
                            break;
                        case '6':
                            stock._u_khistory = meta_khistory_last_id_sh;
                            break;
                        default:
                            stock._u_khistory = meta_khistory_last_id_su;
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
            if (khistory.length) {
                fields_update.push("khistory_to");
                fields_update.push("khistory_from");
                fields_update.push("_u_khistory");
            }
            promise = this.db.update("snapshot", fields_update, update_data);
        } else {
            promise = Promise.resolve(stocks);
        }
        return promise.then(function () {
            if (khistory.length > 0) {
                return this.db.update("khistory", ["code", "date" ], khistory);
            }
            return stocks;
        }.bind(this)).then(function () {
            return stocks;
        }).then(function () {
            if (khistory.length > 0) {
                for (let i = 0; i < stocks.length; i++) {
                    let stock = stocks[i];
                    let code = stock.code;
                    stock.khistory = khistory_map[code];
                }
            }
            if (refresh_view) {
                this.table_init(view_data);
            }
            return stocks;
        }.bind(this));
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

    }
};