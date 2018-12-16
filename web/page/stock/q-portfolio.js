const portfoilio_methods = {
    portfolio_list: function () {
        return axios.post("/cmd/go", {
            "type": "db",
            "cmd": "Keys",
            "args": ["common", "", "portf_*", null],
        }).then(function (resp) {
            let data = util.handle_response(resp);
            for (let i = 0; i < data.length; i++) {
                let name = data[i];
                name = name.replace("portf_", "");
                data[i] = name;
            }
            this.portfolio_names = data.sort();
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
            "args": ["common", portfolio_name, codes_sel, codes_sel, true, 0, null]
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
        return axios.post("/cmd/go", {
            "type": "db",
            "cmd": "Deletes",
            "args": ["common", portfolio, codes, null]
        }).then(function (resp) {
            util.handle_response(resp);
            return this.portfolio_view();
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
            "args": ["common", "", portfolio_name, 1, null]
        }).then(function (resp) {
            let data = util.handle_response(resp);
            let codes = QUtil.keys(data, function (m, k, v) {
                return v;
            });
            return this.stock_get_data_by_code(codes);
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
        return axios.post("/cmd/go", {
            "type": "db",
            "cmd": "Delete",
            "args": ["common", "", key, null]
        }).then(function (resp) {
            util.handle_response(resp);
            this.portfolio.name = "";
            return this.portfolio_list();
        }.bind(this));
    }
};