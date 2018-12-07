
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

var ofields = [{
    "name" : "name",
    "title" : "名字"
}, {
    "name" : "change_rate",
    "title" : "波动"
}, {
    "name" : "now",
    "title" : "价格"
}, {
    "name" : "pb",
    "title" : "PB"
}, {
    "name" : "turnover",
    "title" : "换手",
    "visible" : false
}];


var _columns_default = [
    {
        "name" : "operate",
        "title" : "o",
        "visible" : false,
        "checkbox" : true
        // "formatter2" : function(value, row, index, field) {
        //     return index + "";
        // }
    },
    {
        "name": "code",
        "title": "编码",
        "visible" : false
    },
    {
        "name" : "name",
        "title" : "名字",
        "visible" : true,
        "formatter2" : function(value, row, index, field) {
            return row.name + "<br/>" + row.code + "<br/><pre style='font-size: 0.8em'>" + row._u + "</pre>";
        }
    },
    {
        "name" : "now",
        "title" : "当前",
        "visible" : true,
        "formatter2" : function(value, row, index, field) {
            return value + "<br/><pre style='font-size:0.85em;color:grey;'>" + row.high + "\n" + row.open + "\n" + row.low  + "\n" + row.close +  "</pre>";
        },
        "cellStyle" : function(value, row, index, field) {
            let change_rate = row.change_rate * 1;
            let color = (change_rate > 0) ? "red" : "green"
            return {
                css: { "color": color }
            };
        }
    },
    {
        "name" : "open",
        "title" : "开盘",
        "visible" : false
    },
    {
        "name" : "close",
        "title" : "收盘",
        "visible" : false
    },
    {
        "name" : "swing",
        "title" : "振幅",
        "visible" : false
    },
    {
        "name" : "change",
        "title" : "涨跌金额",
        "sortable" : true,
        "visible" : false,
        formatter : function(value, row, index, field) {
            return row.change + "<br/>" + row.change_rate + "%";
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
        "name" : "change_rate",
        "title" : "涨跌",
        "sortable" : true,
        "visible" : true,
        "formatter2" : function(value, row, index, field) {
            return row.change_rate;
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
        "name" : "vtotal",
        "title" : "总市",
        "sortable" : true,
        "visible" : false
    },
    {
        "name" : "vcir",
        "title" : "流市",
        "sortable" : true,
        "visible" : false
    },
    {
        "name" : "pb", /* 巿淨 */
        "title" : "PB",
        "sortable" : true,
        "visible" : true,
        "formatter2" : function(value, row, index, field){
            return value;
        }
    },
    {
        "name" : "pe", /* 巿盈 */
        "title" : "PE",
        "sortable" : true,
        "visible" : false
    },
    {
        "name" : "turnover",
        "title" : "换手",
        "sortable" : true,
        "visible" : true
    },
    {
        "name" : "appointRate",
        "title" : "委比",
        "sortable" : true,
        "visible" : false,
        "formatter2" : function(value, row, index, field){
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
        "name" : "appointDiff",
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
        "name" : "totalcapital",
        "title" : "总股本",
        "visible" : false
    },
    {
        "name" : "currcapital",
        "title" : "流通股本",
        "visible" : false
    },
    {
        "name" : "max",
        "title" : "最高",
        "visible" : false
    },
    {
        "name" : "min",
        "title" : "最低",
        "visible" : false
    },
    {
        "name" : "vol",
        "title" : "成交量",
        "sortable" : true,
        "visible" : false
    },
    {
        "name" : "amount",
        "title" : "成交金额",
        "sortable" : true,
        "visible" : false
    },
    /*
    {
        "name" : "buy1_n",
        "title" : "#B1",
        "visible" : false
    },
    {
        "name" : "buy1_m",
        "title" : "B1",
        "visible" : false
    },
    {
        "name" : "buy2_n",
        "title" : "#B2",
        "visible" : false
    },
    {
        "name" : "buy2_m",
        "title" : "B2",
        "visible" : false
    },
    {
        "name" : "buy3_n",
        "title" : "#B3",
        "visible" : false
    },
    {
        "name" : "buy3_m",
        "title" : "B3",
        "visible" : false
    },
    {
        "name" : "buy4_n",
        "title" : "#B4"
        ,"visible" : false
    },
    {
        "name" : "buy4_m",
        "title" : "B4",
        "visible" : false
    },
    {
        "name" : "buy5_n",
        "title" : "#B5"
    },
    {
        "name" : "buy5_m",
        "title" : "B5"
    },
    {
        "name" : "sell1_n",
        "title" : "#S1"
    },
    {
        "name" : "sell1_m",
        "title" : "S1"
    },
    {
        "name" : "sell2_n",
        "title" : "#S2"
    },
    {
        "name" : "sell2_m",
        "title" : "S2"
    },
    {
        "name" : "sell3_n",
        "title" : "#S3"
    },
    {
        "name" : "sell3_m",
        "title" : "S3"
    },
    {
        "name" : "sell4_n",
        "title" : "#S4"
    },
    {
        "name" : "sell4_m",
        "title" : "S4"
    },
    {
        "name" : "sell5_n",
        "title" : "#S5"
    },
    {
        "name" : "sell5_m",
        "title" : "S5"
    },
    */
    {
        "name" : "_u",
        "title" : "u",
        "visible" : false
    },
    {
        "name" : "advance",
        "title" : "a",
        "visible" : false,
        "formatter2" : function(value, row, index, field) {
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