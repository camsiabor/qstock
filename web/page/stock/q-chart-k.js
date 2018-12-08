
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
    watch : {

    },
    methods: {
        act : function (action, data, index) {
        },
        chart_render : function () {

            if (!this.rowData || !this.rowData.khistory) {
                return;
            }

            let data = this.rowData.khistory;
            for(let i = 0; i < data.length; i++) {
                let one = data[i];
                let date = one.date;
                let date2 = date.substring(0, 4) + "-" + date.substring(4, 6) + "-" + date.substring(6, 8);
                one.date2 = date2;
            }

            this.chart && this.chart.destory();
            let ds = new DataSet({
                state: {
                    start: 20181101,
                    end: 20181208
                }
            });

            let dv = ds.createView();
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
            this.chart = new G2.Chart({
                container: this.cid,
                forceFit: true,
                height: window.innerHeight / 3,
                animate: false
                // padding: [10, 40, 40, 40]
            });
            this.chart.source(dv, {
                'date2': {
                    type: 'timeCat',
                    nice: false,
                    range: [0, 1]
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
            // chart.legend({
            //     offset: 20
            // });
            this.chart.tooltip({
                showTitle: false,
                itemTpl: '<li data-index={index}>' + '<span style="background-color:{color};" class="g2-tooltip-marker"></span>' + '{name}{value}</li>'
            });

            let kView = this.chart.view({
                end: {
                    x: 1,
                    y: 1
                }
            });
            kView.source(dv);
            kView.schema().position('date2*range').color('trend', function(val) {
                if (val === '上涨') {
                    return '#f04864';
                }
                if (val === '下跌') {
                    return '#2fc25b';
                }
            }).shape('candle').tooltip('date2*open*close*high*low', function(date2, open, close, high, low) {
                return {
                    name: date2,
                    value: '<br><span style="padding-left: 16px">开盘价：' + open + '</span><br/>' + '<span style="padding-left: 16px">收盘价：' + end + '</span><br/>' + '<span style="padding-left: 16px">最高价：' + high + '</span><br/>' + '<span style="padding-left: 16px">最低价：' + low + '</span>'
                };
            });
            this.chart.render();

        }
    },
    mounted : function () {
        this.chart_render();
        // setTimeout(function () {
        //     this.chart_render();
        // }.bind(this), 100);
    }
});
