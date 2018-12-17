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
        let hash = "";
        let script = this.editor.getValue().trim();
        if (window.location.href.indexOf('nohash') <= 0) {
            hash = md5(script);
            if (this.hash[hash]) {
                script = "";
            }  else {
                this.hash[hash] = hash
            }
        }
        this.console.text = "";
        return axios.post("/cmd/go", {
            type : 'lua',
            cmd : 'run',
            hash : hash,
            script : script
        }).then(this.stock_get_data_by_code)
    },

    script_test: function () { // eJyqVkrOT0lVsjLQUUpJLElUsgIJ5BWX5iKL5ZXm5OgopaQmlaYrWZUUlabW1gICAAD///qyEpY=
        let hash = "";
        let script = this.editor.getValue().trim();
        if (window.location.href.indexOf('nohash') <= 0) {
            hash = md5(script);
            if (this.hash[hash]) {
                script = "";
            }  else {
                this.hash[hash] = hash;
            }
        }
        this.console.text = "";
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
        name = name || this.portfolio.name;
        this.setting.params.last = name;
        this.params_get(name);
    },

    params_update : function (name) {
        if (!name) {
            name = this.params.name;
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
        if (!name) {
            name = this.params.name;
        }
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
        this.params.list.push({
            "key" : "key",
            "alias" : "alias",
            "value" : "value",
            "desc" : "desc"
        });
    },
    
    params_list_delete : function () {

    }

};