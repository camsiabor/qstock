
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


const datamock = [{
    code: "greetings",
    name: "redis",
    now: 1, open: 1, min: 1, max: 1, close: 1,
    change_reate: 1
},{
    code: "greetings",
    name: "lua",
    now: 3, open: 1, min: 1, max: 1, close: 1,
    change_reate: 1
}];


const cssmock = {
    table: {
        tableWrapper: '',
        tableHeaderClass: 'mb-0',
        tableBodyClass: 'mb-0',
        tableClass: 'table table-bordered table-hover',
        loadingClass: 'loading',
        ascendingIcon: 'fa fa-chevron-circle-up',
        descendingIcon: 'fa fa-chevron-circle-down',
        handleIcon: 'fa-chrome',
        handleIcon2: 'fa fa-bars text-secondary',

        ascendingClass: 'sorted-asc',
        descendingClass: 'sorted-desc',
        sortableIcon: 'fa fa-sort',
        detailRowClass: 'vuetable-detail-row',

        renderIcon(classes, options) {
            return `<i class="${classes.join(' ')}"></span>`
        }
    },
    pagination: {
        wrapperClass: "pagination pull-right",
        activeClass: "btn btn-outline-info",
        disabledClass: "disabled",
        pageClass: "btn page-item",
        linkClass: "page-link",
        icons: {
            first: "",
            prev: "",
            next: "",
            last: ""
        }
    }
};

const _columns_default = [
    {
        "name" : "operate",
        "title" : "o",
        "visible" : false,
        "checkbox" : true
        // "callback" : function(value, row, col, vuetable) {
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
        "callback": function (value, row, col, vuetable) {
            return row.name + "<br/>" + row.code + "<br/><pre style='font-size: 0.8em'>" + row._u + "</pre>";
        }
    },
    {
        "name" : "now",
        "title" : "当前",
        "visible" : true,
        "callback" : function(value, row, col, vuetable) {
            let now_color = QUtil.stock_color(row.now * 1 - row.open * 1);
            let open_color = QUtil.stock_color(row.open * 1 - row.close * 1);
            let low_color = QUtil.stock_color(row.low * 1 - row.close * 1);
            let high_color = QUtil.stock_color(row.high * 1 - row.close * 1);
            let html = [];
            html.push("<div class='s-bold' style='color:" + now_color + "'>" + value + "</div>");
            html.push("<div class='s-bold s-tiny' style='color:" + open_color + "'>"  + row.close + " -> " + row.open + "</div>");
            html.push("<div class='s-bold s-tiny'>");
            html.push("<span style='color:" + low_color + "'>" + row.low + "</span>");
            html.push(" - ")
            html.push("<span style='color:" + high_color + "'>" + row.high + "</span>");
            html.push("</div>");
            return html.join("");
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
        "callback" : function(value, row, col, vuetable) {
            row.change_rate = row.change_rate * 1;
            let color = QUtil.stock_color(row.change_rate);
            return "<span style='color:'" + color + "'>" + row.change_rate + "</span>";
        }
    },
    {
        "name" : "change_rate",
        "title" : "涨跌",
        "sortField" : "change_rate",
        "visible" : true,
        "callback" : function(value, row, col, vuetable) {
            row.change_rate = row.change_rate * 1;
            let color = QUtil.stock_color(row.change_rate);
            return "<span class='s-bold' style='color:" + color + "'>" + row.change_rate + " %</span>";
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
        "callback" : function(value, row, col, vuetable){
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
        "callback" : function(value, row, col, vuetable){
            return row.appointRate + "%<br/>" + row.appointDiff;
        },
        "cellStyle" : function(value, row, col, vuetable) {
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
        "cellStyle" : function(value, row, col, vuetable) {
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
        "callback" : function(value, row, col, vuetable) {
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