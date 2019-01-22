const script_methods = {


    script_setting: function(opts) {
        this.script_setting_opts = opts;
        $('#div_script_setting').modal('toggle');
    },

    script_list: function () {
        return axios.post("/script/list").then(function (json) {
            let names = util.handle_response(json, this.console, "");
            this.script_names = names.sort().reverse();
        }.bind(this)).catch(util.handle_error.bind(this));
    },

    script_group_list : function() {
        return axios.post("/cmd/go", {
            "type": "db",
            "cmd": "Keys",
            "args": ["common", "script_group", "portf_*", null],
        }).then(function (resp) {
            let data = util.handle_response(resp);
            for (let i = 0; i < data.length; i++) {
                // let name = data[i];
                // name = name.replace("portf_", "");
                // data[i] = name;
            }
            // this.portfolio_names = data.sort();
        }.bind(this));
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
            this.script_query();

            if (this.timer_script_save) {
                clearTimeout(this.timer_script_save);
            }
            this.timer_script_save = setTimeout(this.script_save.bind(this), 10 * 60 * 1000);

        }.bind(this)).catch(util.handle_error.bind(this))
    },

    script_save: function () {
        if (!this.script.name) {
            util.popover("#button_script_save", "需要脚本名字", "bottom");
            return;
        }
        this.setting.script.last = this.script.name;
        this.script.script = this.editor.getValue().trim();
        return axios.post("/script/update", this.script).then(function (resp) {
            util.handle_response(resp, this.console, "script saved @ " + this.script.name)
            util.popover("#button_script_save", "保存成功", "bottom");
            this.script_list();
            return resp
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


    script_query: function (mode, carrayscript) {

        let script = this.editor.getValue().trim();
        if (script.length === 0) {
            alert("need script!");
            return;
        }
        mode = mode  || this.setting.mode;

        let nohash = window.location.href.indexOf('nohash') > 0
        let hash = md5(script);
        if (!this.hash[hash]) {
            this.hash[hash] = hash
        }
        if (!carrayscript) {
            script = hash ? "" : script;
        }
        this.console.text = "";

        return axios.post("/cmd/go", {
            type : 'lua',
            cmd : 'run',
            hash : nohash ? "" : hash,
            mode : mode,
            script : script,
            params : this.params,
            name : this.script.name,
        }, {
            timeout: this.setting.script.timeout || 300000
        }).then(function (resp) {
            if (resp.data.code === 404) {
                return this.script_query(mode, true);
            }
            let data = util.handle_response(resp);
            if (mode === "debug") {

                if (typeof data === 'object') {
                    data = JSON.stringify(data, null, 2);
                }
                this.console.text = data;
                return data;
            } else {
                data.refresh_view = true;
                return this.stock_get_data_by_code(data);
            }
        }.bind(this));


    },


    params_list: function() {
        return axios.post("/cmd/go", {
            "type": "db",
            "cmd": "Keys",
            "args": [ "common", "params", "*", null ],
        }).then(function (resp) {
            let data = util.handle_response(resp);
            this.params_names = data.sort();
        }.bind(this));
    },

    params_setting : function () {
        if (!this.params.name) {
            alert("请输入参数名字");
            return;
        }
        $('#div_params_setting').modal('toggle');
    },

    params_get: function(name) {
        if (!name) {
            name = this.params.name;
        }
        if (!name) {
            return;
        }
        return axios.post("/cmd/go", {
            "type": "db",
            "cmd": "Get",
            "args": ["common", "params", name, 1, null]
        }).then(function (resp) {
            let data = util.handle_response(resp);
            if (data) {
                this.params = data;
            }
        }.bind(this));
    },

    params_select: function(name) {
        name = name || this.params.name.trim();
        if (name) {
            this.setting.params.last = name;
            this.params_get(name);
        }
    },

    params_update : function (name) {
        if (!name) {
            name = this.params.name.trim();
        }
        if (!name) {
            alert("需要参数名字");
            return;
        }
        axios.post("/cmd/go", {
            "type": "db",
            "cmd": "Update",
            "args": ["common", "params", name, this.params, true, 1, null]
        }).then(function (resp) {
            let msg = "保存参数 " + this.params.name + " 成功"
            util.handle_response(resp, this.console, msg);
            this.params_list();
        }.bind(this));
    },

    params_delete : function (name) {
        name = name || this.params.name.trim();
        if (!confirm("sure to delete params? " + name)) {
            return;
        }
        return axios.post("/cmd/go", {
            "type": "db",
            "cmd": "Delete",
            "args": ["common", "params", name, null]
        }).then(function (resp) {
            util.handle_response(resp);
            this.params.name = "";
            return this.params_list();
        }.bind(this));
    },

    params_list_add : function () {
        if (!this.params) {
            this.params = {};
        }
        if (!this.params.list) {
            this.params.list = [];
        }
        let len = this.params.list.length;
        this.params.list.push({
            "key" : "key" + len,
            "alias" : "key" + len,
            "value" : "val" + len,
            "expression" : ""
        });
    },
    
    params_list_delete : function (index) {
        let head = this.params.list.slice(0, index);
        let tail = this.params.list.slice(index + 1);
        this.params.list = head.concat(tail);
    }

};