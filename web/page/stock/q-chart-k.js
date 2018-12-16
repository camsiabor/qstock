
/* ===================== vue component vueable-chart ============================ */


Vue.component('vuetable-chart', {
    template : "<div v-bind:id='cid' style='width: 99%; height: 99%;'></div>",
    props: {
        rowData: {
            type: Object,
            required: true
        },
        rowIndex: {
            type: Number
        }
    },
    computed : {
        cid : function () {
            return 'vt-chart-' + this.rowIndex;
        }
    },
    methods: {
        chart_render : function (stocks_map) {

            if (this.timer_render) {
                clearTimeout(this.timer_render);
            }

            if (this.chart) {
                try {
                    this.chart.destroy();
                } catch (e) {}
                this.chart = null;
            }

            if (!this.rowData || !this.rowData.code) {
                console.log(this.cid, "no row data", this.rowData);
                return;
            }
            let code = this.rowData.code;
            let stock = stocks_map[code];
            if (!stock) {
                console.log(this.cid, "no stock data", code, stocks_map);
                return;
            }
            let data = stock.khistory;
            if (!data || !data.length) {
                console.log(this.cid, "khistory null", stock, data);
                this.timer_render = setTimeout(function () {
                    this.chart_render(stocks_map);
                }.bind(this), 2000);
                return;
            }

            let nowdate = QUtil.date_format(new Date(), "");
            let data_newest = QUtil.array_most(data, function (most, one) {
                return one.date * 1 > most.date * 1;
            });

            let has_current = QUtil.map_is_same_by_field_val(stock, data_newest, [ "open", "high", "low" ], 0.01);
            if (!has_current) {
                let clone = QUtil.map_clone(stock);
                clone.date = nowdate;
                clone.close = stock.now;
                data.push(clone);
            }

            let kagi = stocks_map["kagi"] || {};
            let kagi_count = kagi.count || 16;
            let kagi_height = kagi.height * 1 || 100;
            let kagi_scale_y = kagi.scale_y || 2;
            if (!kagi.height) {
                kagi_height = (window.innerHeight / 3);
            }
            if (kagi.height < 0) {
                kagi_height = (window.innerHeight / (-kagi_height));
            }
            kagi_height = Math.floor(kagi_height);
            let kagi_width = kagi.width * 1 ||  0;

            if (isNaN(kagi_width)) {
                kagi_width = 0;
            }
            if (isNaN(kagi_height)) {
                kagi_height = 100;
            }

            for(let i = 0; i < data.length; i++) {
                let one = data[i];
                let date = one.date;
                let date2 = date.substring(0, 4) + "-" + date.substring(4, 6) + "-" + date.substring(6, 8);
                one.date2 = date2;
            }

            // console.log(this.cid, data, this.chart);

            let time_end = new Date();
            let time_start = QUtil.date_add_day(time_end, -kagi_count);
            let time_end_str = QUtil.date_format(time_end, "");
            let time_start_str = QUtil.date_format(time_start, "");
;
            let ds = new DataSet({
                state: {
                    end: time_end_str * 1,
                    start: time_start_str * 1
                }
            });

            let dv = ds.createView();
            try {
                dv.source(data).transform({
                    type: 'filter',
                    callback: function callback(obj) {
                        let date = obj.date * 1;
                        if (isNaN(date)) {
                            return false;
                        }
                        return date >= ds.state.start && date <= ds.state.end;
                    }
                }).transform({
                    type: 'map',
                    callback: function callback(obj) {
                        obj.trend = obj.open <= obj.close ? '上涨' : '下跌';
                        obj.range = [obj.open, obj.close, obj.high, obj.low];
                        return obj;
                    }
                });
            } catch (e) {
                console.error("[kagi]", e, code, stock, data);
                return;
            }


            let chart_opt  = {
                container: this.cid,
                animate: false,
                height: kagi_height
            };
            if (kagi_width === 0) {
                chart_opt.forceFit = true;
            } else if (kagi_width > 0) {
                chart_opt.width = kagi_width;
            } else {
                chart_opt.width = window.innerWidth / (-kagi_width);
            }

            this.chart = new G2.Chart(chart_opt);
            this.chart.source(dv, {
                'date2': {
                    type: 'timeCat',
                    nice: false,
                    range: [0, 1],
                    tick : 1,
                    tickInterval: 2 * 24 * 60 * 60 * 1000
                },
                trend: {
                    values: ['上涨', '下跌']
                },
                'vol': {
                    alias: '成交量'
                },
                'open': {
                    alias: '开盘价'
                },
                'close': {
                    alias: '收盘价'
                },
                'high': {
                    alias: '最高价'
                },
                'low': {
                    alias: '最低价'
                },
                'range': {
                    alias: '股票价格'
                }
            });

            // this.chart.axis('date2', {
            //     label: {
            //         formatter: val => {
            //             return val.substring(5);
            //         }
            //     }
            // });

            this.chart.axis('date2', {
                label: null
            });

            this.chart.axis('range', {
                label: null
            });

            this.chart.scale('x', {
                tickCount: 1
            });

            this.chart.legend('trend', false); // 不显示 cut 字段对应的图例
            // this.chart.legend('trend', {
            //     // offset : 30,
            //     position : 'right-center'
            // }); // 不显示 cut 字段对应的图例
            this.chart.tooltip({
                showTitle: false,
                itemTpl: '<li data-index={index}>' + '<span style="background-color:{color};" class="g2-tooltip-marker"></span>' + '{name}{value}</li>'
            });

            let kView = this.chart.view({
                end: {
                    x: 0.9,
                    y: kagi_scale_y
                }
            });
            kView.source(dv);

            kView.schema().position('date2*range').color('trend', function(val) {
                if (val === '上涨') {
                    return QUtil.COLOR_UP;
                }
                if (val === '下跌') {
                    return QUtil.COLOR_DOWN;
                }
            }).shape('candle').tooltip('date2*open*close*high*low*change_rate', function(date2, open, close, high, low, change_rate) {
                let html = [];
                let color = QUtil.stock_color(change_rate);
                html.push('<br><span style="padding-left: 1px">open ');
                html.push(open);
                html.push('</span><br/>');
                html.push('<span style="padding-left: 1px">close ');
                html.push(close);
                html.push('</span><br/>');
                html.push('<span style="padding-left: 1px">high ');
                html.push(high);
                html.push('</span><br/>');
                html.push('<span style="padding-left: 1px">low ');
                html.push(low);
                html.push('</span><br/>');
                html.push('<span style="padding-left: 1px;color:' + color + '">rate ');
                html.push((change_rate+"").substring(0, 4) + "%");
                html.push('</span>');
                return {
                    name: date2.substring(5),
                    value: html.join("")
                };
            });


            this.chart.render();

        }
    }
});

/* ===================== vue component vueable-actions ============================ */

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