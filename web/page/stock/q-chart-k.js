
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
            this.chart && this.chart.destory();
            const data = [
                { genre: 'Sports', sold: 275 },
                { genre: 'Strategy', sold: 115 },
                { genre: 'Action', sold: 120 },
                { genre: 'Shooter', sold: 350 },
                { genre: 'Other', sold: 150 }
            ];
            this.chart = new G2.Chart({
                container: this.cid,
                // width: 100,
                height: 100,
                forceFit: true,
            });
            this.chart.source(data);
            this.chart.interval().position('genre*sold').color('genre')
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
