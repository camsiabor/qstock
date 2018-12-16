const script_methods = {
    script_list: function () {
        return axios.post("/script/list").then(function (json) {
            let names = util.handle_response(json, this.console, "");
            this.script_names = names.sort();
        }.bind(this)).catch(util.handle_error.bind(this))
    },

    script_select: function (name) {
        this.script.name = name;
        this.setting.script.last = name;
        return axios.post("/script/get", {
            name: name
        }).then(function (resp) {
            let info = util.handle_response(resp);
            this.script.name = info.name;
            this.script.script = info.script;
            this.editor.setValue(this.script.script);
            this.editor.clearSelection();
            if (this.setting.mode === "query") {
                this.script_query();
            }
        }.bind(this)).catch(util.handle_error.bind(this))
    },

    script_save: function () {
        if (!this.script.name) {
            util.popover("#button_script_save", "需要脚本名字", "bottom");
            return;
        }
        this.setting.script.last = this.script.name;
        this.script.script = this.editor.getValue().trim();

        axios.post("/script/update", this.script).then(function (resp) {
            util.handle_response(resp, this.console, "script saved @ " + this.script.name)
            util.popover("#button_script_save", "保存成功", "bottom");
            this.script_list();
        }.bind(this)).catch(util.handle_error.bind(this));
    },

    script_delete: function () {
        if (!confirm("sure to delete? " + this.script.name)) {
            return;
        }
        axios.post("/script/delete", {
            name: this.script.name
        }).then(function (resp) {
            util.handle_response(resp, this.console, "script deleted @ " + this.script.name)
            this.script.name = "---";
            this.script_list();
        }.bind(this)).catch(util.handle_error.bind(this));
    },


    script_query: function () {
        this.console.text = "";
        let script = this.editor.getValue().trim();
        let hash = md5(script);
        return axios.post("/cmd/go", {
            type : 'lua',
            cmd : 'run',
            hash : hash,
            script : script
        }).then(this.stock_get_data_by_code)
        /*
         return axios.post("/cmd/query", {
            type : 'db',
            cmd : 'run',
            script : script
        }).then(this.stock_get_data_by_code);
        */
    },

    script_test: function () {
        let script = this.editor.getValue().trim();
        let hash = "";
        if (window.location.href.indexOf('nohash') <= 0) {
            hash = md5(script);
        }
        return axios.post("/cmd/go", {
            type : 'lua',
            cmd : 'run',
            hash : hash,
            debug : true,
            script : script
        }).then(function (resp) {
            let data = util.handle_response(resp);
            if (typeof data === 'object') {
                data = JSON.stringify(data, null, 2);
            }
            this.console.text = data;
            return data;
        }.bind(this))
    },

    params_list: function() {
        return axios.post("/cmd/go", {
            "type": "db",
            "cmd": "Keys",
            "args": ["common", "", "params", null ],
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

    params_select: function() {

    },

    params_setting : function () {

    },

    params_save : function () {

    },

    params_delete : function () {

    }

};