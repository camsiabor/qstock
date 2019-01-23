const script_methods = {

    script_list: function (type) {
        return axios.post("/script/list", { type : type }).then(function (json) {
            let data = util.handle_response(json, this.console, "");

            if (type.indexOf("group") >= 0) {

                let script_group;
                if (type.indexOf(",") >= 0) {
                    script_group = data["script_group"][0];
                } else {
                    script_group = data[0];
                }
                script_group = script_group || {
                    "id" : "system",  "label" : "system"
                };

                if (script_group.tree) {
                    if (typeof script_group.tree === "string") {
                        script_group.tree = JSON.parse(script_group.tree);
                    }
                } else {
                    script_group.tree = [];
                }
                let result = QUtil.tree_locate(script_group.tree, { id : "all" },  {
                    depth_limit : 0
                });
                let all = result && result.target;
                if (!all) {
                    script_group.tree.push({ id : "all", label : "all", children : [] });
                }
                script_group.id = script_group.id || "system";
                script_group.name = script_group.name || this.script_group.id;
                this.script_group = script_group;
                this.script_group_only.tree = QUtil.tree_clone(this.script_group.tree, {
                    cloner : function (tree, node, opts) {
                        if (node.children) {
                            return QUtil.map_clone(node);
                        }
                    }
                });
            }

            if (type.indexOf("script") >= 0) {
                let script_names;
                if (type.indexOf(",") >= 0) {
                    script_names = data["script_names"];
                } else {
                    script_names = data;
                }
                this.script_names = script_names.sort().reverse();

                let tree = this.script_group.tree;
                let result = QUtil.tree_locate(this.script_group.tree, {id: "all"}, {
                    depth_limit: 0
                });
                let all = result && result.target;
                if (all) {
                    all.children = [];
                } else {
                    all = {id: "all", label: "all", children: []};
                    tree.push(all);
                }
                for (let i = 0, n = this.script_names.length; i < n; i++) {
                    let name = this.script_names[i];
                    all.children.push({id: name, label: name});
                }
            }

        }.bind(this)).catch(util.handle_error.bind(this));
    },

    script_select: function (noderaw, id, node) {

        if (!node && noderaw) { /* active select */
            node = this.$refs.tree_script.getNode(noderaw.id);
            if (node) {
                this.$refs.tree_script.select(node);
                return;
            }
        }

        /*
        if (node.children) {
            this.$refs.tree_script.clear();
            return;
        }
        */

        this.setting.script.last = this.script.name = node.id;
        return axios.post("/script/get", {
            name: this.script.name
        }).then(function (resp) {
            let info = util.handle_response(resp);
            this.script.script = info.script;
            this.editor.setValue(this.script.script);
            this.editor.clearSelection();
            this.script_query();

            if (this.timer_script_save) {
                clearTimeout(this.timer_script_save);
            }
            this.timer_script_save = setTimeout(
            function() {
                this.script_save( { type : "script" } );
            }.bind(this), 10 * 60 * 1000);

            this.config_persist();
        }.bind(this)).catch(util.handle_error.bind(this))
    },




    script_save: function (opts) {
        let type = opts.type;
        let name = opts.name;
        if (name) {
            name = name.trim();
        }
        if (type === "script") {
            if (!name) {
                name = this.script.name;
            }
            if (!name) {
                util.popover("#button_script_save", "需要名字", "bottom");
                return;
            }
            this.setting.script.last = this.script.name = name;
            this.script.script = this.editor.getValue().trim();
            return axios.post("/script/update", this.script).then(function (resp) {
                util.handle_response(resp, this.console, "script saved @ " + this.script.name);
                util.popover("#button_script_save", "保存成功", "bottom");
                this.script_list(opts.type);
                return resp
            }.bind(this)).catch(util.handle_error.bind(this));
        } else {
            if (name) {
                name = name.trim();
                this.script_group.tree.push({
                    id : "g" + new Date().getTime() + (Math.random() + "").substring(0, 8),
                    label : name,
                    children : []
                });
            }
            let script_group_obj = QUtil.map_clone(this.script_group);
            script_group_obj.tree = JSON.stringify(this.script_group.tree);
            return axios.post("/cmd/go", {
                "type": "db",
                "cmd": "Update",
                "args": ["common", "script_group", script_group_obj.id, script_group_obj, true, 1, null]
            }).then(function (resp) {
                util.handle_response(resp, this.console, "");
                this.script_setting_opts.msg = "保存群组成功";
                this.script_list(opts.type);
            }.bind(this));
        }
    },


    script_add: function(opts) {
        let name = prompt("请输入名字");
        if (name) {
            name = name.trim();
        }
        if (!name) {
            return;
        }
        let parent;
        let selected = this.script_setting_opts.node;
        if (selected) {
            parent = selected.children;
        } else {
            if (opts.type === "script") {
                let result = QUtil.tree_locate(this.script_group.tree, { id : "all" },  {
                    depth_limit : 0
                });
                parent = result.target;
                parent = parent.children;
            } else {
                parent = this.script_group.tree;
            }
        }
        let node = { label : name };
        if (opts.type === "script") {
            node.id = name;
            if (!parent) {
                alert("需要分组节点");
                return;
            }
        } else {
            node.id = "g" + (new Date().getTime()) + (Math.random() + "").substring(0, 8);
            node.children = [];
        }
        let existing;
        if (opts.type === "script") {
            existing = QUtil.tree_locate(this.script_group.tree, { field_id : "id", id : node.id }, {});
        } else {
            existing = QUtil.tree_locate(parent, { field_id : "label", id : node.name }, {});
        }

        if (existing) {
            alert("节点已存在: " + existing.target.label);
            return;
        }
        parent.push(node);
        if (opts.type === "script") {
            this.script_save( { type : "script", name : name } );
        }
        this.script_save( { type : "group" } );
    },

    script_delete: function (opts) {
        if (opts.type === "script") {
            if (!confirm("sure to delete? " + this.script.name)) {
                return;
            }
            axios.post("/script/delete", {
                name: this.script.name
            }).then(function (resp) {
                util.handle_response(resp, this.console, "script deleted @ " + this.script.name)
                this.script.name = null;
                this.script_list( { type : "script,group" } );
            }.bind(this)).catch(util.handle_error.bind(this));
        } else {

        }

    },


    script_query: function (mode, carrayscript) {
        let script = this.editor.getValue().trim();
        if (script.length === 0) {
            this.console.text = "script content is empty";
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
                if (data) {
                    data.refresh_view = true;
                    return this.stock_get_data_by_code(data);
                }
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
    },

    /* ======================================== script setting ====================================================== */
    script_setting: function(type) {

        if (type === "script") {
            this.script_setting_opts = {
                type : type,
                title : "脚本设置",
                tree_src_value : null,
                tree_des_value : null,
                tree_src_desc : "选择脚本",
                tree_des_desc : "选择分组",
                tree_src_value_consists_of : "LEAF_PRIORITY",
                tree_src_select_consists_of : "LEAF",
                tree_src_display_consists_of : "ALL",
                tree_src : this.script_group.tree,
                tree_des : this.script_group.tree
            };
        } else {
            this.script_setting_opts = {
                type : type,
                title : "脚本分组设置",
                tree_src_value : null,
                tree_des_value : null,
                tree_src_desc : "选择来源分组",
                tree_des_desc : "选择目标分组",
                tree_src_display_consists_of : "BRANCH",
                tree_src : this.script_group.tree,
                tree_des : this.script_group.tree
            };
        }

        $('#div_script_setting').modal('toggle');
    },

    script_group_tree_src_select : function(node, id) {
        this.script_setting_opts.tree_src_node = node;
    },

    script_group_tree_des_select : function(node, id) {
        this.script_setting_opts.tree_des_node = node;
    },

    script_group_move : function() {
        let tree_script_src = this.$refs.tree_script_src;
        let tree_script_des = this.$refs.tree_script_des;

        let nodes_src = tree_script_src.selectedNodes;
        let nodes_des = tree_script_des.selectedNodes;

        if (nodes_src.length === 0 || nodes_des.length === 0) {
            alert("需要选择來源及目标");
            return;
        }


    },

    script_group_copy : function() {

    }

};