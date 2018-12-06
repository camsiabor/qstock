
/*
var MyComponent= Vue.extend({
    template: '<a style="color:#07bb49;" v-on:click="world('+"'"+id+"'"+')">删除</a>',
    methods:{
        world:function(id) {
            alert(this.data);
        }
    }
});
*/


var _columns_default = [
    {
        "field" : "operate",
        "title" : "o",
        "visible" : true,
        "checkbox" : true
        // "formatter" : function(value, row, index, field) {
        //     return index + "";
        // }
    },
    {
        "field": "code",
        "title": "编码",
        "visible" : false
    },
    {
        "field" : "name",
        "title" : "名字",
        "visible" : true,
        "formatter" : function(value, row, index, field) {
            return row.name + "<br/>" + row.code + "<br/><pre style='font-size: 0.8em'>" + row._u + "</pre>";
        }
    },
    {
        "field" : "nowPrice",
        "title" : "当前",
        "visible" : true,
        "formatter" : function(value, row, index, field) {
            return value + "<br/><pre style='font-size:0.85em;color:grey;'>" + row.todayMax + "\n" + row.openPrice + "\n" + row.todayMin + "\n" + row.swing + "%</pre>";
        },
        "cellStyle" : function(value, row, index, field) {
            let diff_money = row.diff_money * 1;
            let color = (diff_money > 0) ? "red" : "green"
            return {
                css: { "color": color }
            };
        }
    },
    {
        "field" : "openPrice",
        "title" : "开盘",
        "visible" : false
    },
    {
        "field" : "closePrice",
        "title" : "收盘",
        "visible" : false
    },
    {
        "field" : "swing",
        "title" : "振幅"
    },
    {
        "field" : "diff_money",
        "title" : "涨跌金额",
        "sortable" : true,
        "visible" : false,
        formatter : function(value, row, index, field) {
            return row.diff_money + "<br/>" + row.diff_rate + "%";
        },
        "cellStyle" : function(value, row, index, field) {
            value = value * 1;
            let color = (value > 0) ? "red" : "green"
            return {
                css: { "color": color }
            };
        }
    },
    {
        "field" : "diff_rate",
        "title" : "涨跌",
        "sortable" : true,
        "visible" : true,
        "formatter" : function(value, row, index, field) {
            return row.diff_rate + "%<br/>" + row.diff_money;
        },
        "cellStyle" : function(value, row, index, field) {
            value = value * 1;
            let color = (value > 0) ? "red" : "green"
            return {
                css: { "color": color }
            };
        }
    },
    {
        "field" : "all_value",
        "title" : "总市",
        "sortable" : true,
        "visible" : true
    },
    {
        "field" : "circulation_value",
        "title" : "流市",
        "sortable" : true,
        "visible" : false
    },
    {
        "field" : "pb", /* 巿淨 */
        "title" : "PB",
        "sortable" : true,
        "visible" : true,
        "formatter" : function(value, row, index, field){
            return value;
        }
    },
    {
        "field" : "pe", /* 巿盈 */
        "title" : "PE",
        "sortable" : true,
        "visible" : true
    },
    {
        "field" : "turnover",
        "title" : "换手",
        "sortable" : true,
        "visible" : true
    },
    {
        "field" : "appointRate",
        "title" : "委比",
        "sortable" : true,
        "visible" : true,
        "formatter" : function(value, row, index, field){
            return row.appointRate + "%<br/>" + row.appointDiff;
        },
        "cellStyle" : function(value, row, index, field) {
            value = value * 1;
            let color = (value > 0) ? "red" : "green"
            return {
                css: { "color": color }
            };
        }
    },
    {
        "field" : "appointDiff",
        "title" : "委差",
        "sortable" : true,
        "visible" : false,
        "cellStyle" : function(value, row, index, field) {
            value = value * 1;
            let color = (value > 0) ? "red" : "green"
            return {
                css: { "color": color }
            };
        }
    },
    {
        "field" : "totalcapital",
        "title" : "总股本",
        "visible" : false
    },
    {
        "field" : "currcapital",
        "title" : "流通股本",
        "visible" : false
    },
    {
        "field" : "todayMax",
        "title" : "最高"
    },
    {
        "field" : "todayMin",
        "title" : "最低"
    },
    {
        "field" : "tradeNum",
        "title" : "成交量",
        "sortable" : true,
        "visible" : false
    },
    {
        "field" : "tradeAmount",
        "title" : "成交金额",
        "sortable" : true,
        "visible" : false
    },
    {
        "field" : "buy1_n",
        "title" : "#B1"
    },
    {
        "field" : "buy1_m",
        "title" : "B1"
    },
    {
        "field" : "buy2_n",
        "title" : "#B2"
    },
    {
        "field" : "buy2_m",
        "title" : "B2"
    },
    {
        "field" : "buy3_n",
        "title" : "#B3"
    },
    {
        "field" : "buy3_m",
        "title" : "B3"
    },
    {
        "field" : "buy4_n",
        "title" : "#B4"
    },
    {
        "field" : "buy4_m",
        "title" : "B4"
    },
    {
        "field" : "buy5_n",
        "title" : "#B5"
    },
    {
        "field" : "buy5_m",
        "title" : "B5"
    },
    {
        "field" : "sell1_n",
        "title" : "#S1"
    },
    {
        "field" : "sell1_m",
        "title" : "S1"
    },
    {
        "field" : "sell2_n",
        "title" : "#S2"
    },
    {
        "field" : "sell2_m",
        "title" : "S2"
    },
    {
        "field" : "sell3_n",
        "title" : "#S3"
    },
    {
        "field" : "sell3_m",
        "title" : "S3"
    },
    {
        "field" : "sell4_n",
        "title" : "#S4"
    },
    {
        "field" : "sell4_m",
        "title" : "S4"
    },
    {
        "field" : "sell5_n",
        "title" : "#S5"
    },
    {
        "field" : "sell5_m",
        "title" : "S5"
    },
    {
        "field" : "_u",
        "title" : "u",
        "visible" : false
    },
    {
        "field" : "advance",
        "title" : "a",
        "visible" : false,
        "formatter" : function(value, row, index, field) {
            return '<input type="button" value="移出" class="btn btn-outline-secondary"  v-on:click="portfolio_unadd()" />'
        }
    }
];

var _columns_default_map = { };
for(let i = 0; i < _columns_default.length; i++) {
    let one = _columns_default[i];
    let key = one.field;
    _columns_default_map[key] = one;
}