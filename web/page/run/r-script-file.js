const script_file_methods = {

    script_file_list: function (opts) {
        opts = opts || this.setting.locate;
        let path = opts.path || ".";
        opts.category = opts.category || "lua";
        return axios.post("/os/file/list", opts).then(function (json) {
            let data = util.handle_response(json, this.console, "");
            for (let i = 0; i < data.length; i++) {
                let one = data[i];
                one.id = path + "/" + one.name;
                one.label = one.name;
                if (one.isdir) {
                    one.children = null;
                }
            }
            if (opts.node) {

            } else {
                this.script_file_group.tree = data;
            }
        }.bind(this)).catch(util.handle_error.bind(this));
    },

    script_file_sublist : function(act) {

        if (act.action !== "LOAD_CHILDREN_OPTIONS") {
            return;
        }
        if (act.parentNode.children && act.parentNode.children.length > 0) {
            return;
        }
        let vroot = this.$root;
        let opts = {};
        opts.path = act.parentNode.id;
        opts.category = vroot.setting.locate.category;
        return axios.post("/os/file/list", opts).then(function (json) {
            let data = util.handle_response(json, vroot.console, "");
            act.parentNode.children = [];
            for (let i = 0; i < data.length; i++) {
                let one = data[i];
                one.id = act.parentNode.id + "/" + one.name;
                one.label = one.name;
                if (one.isdir) {
                    one.children = null;
                }
                act.parentNode.children.push(one);
            }
            act.callback();
        }.bind(this)).catch(function (e) {
            vroot.console.text = e;
            console.error("[load script tree node fail]", e);
            act.callback(e);
        });

    },

    script_file_select: function (noderaw, id, node) {

        if (!node) { /* active select */
            if (noderaw) {
                node = this.$refs.tree_script_file.getNode(noderaw.id);
            } else {
                node = this.$refs.tree_script_file.selectedNodes[0];
            }
            if (node) {
                this.$refs.tree_script_file.select(node);
            }
            return;
        }

        this.setting.script_file.last = this.script_file.name = node.id;
        return axios.post("/os/file/text", {
            path: node.id,
            category: this.setting.locate.category
        }).then(function (resp) {
            let text = util.handle_response(resp);
            this.script_file.script = text;
            this.editor.setValue(this.script_file.script);
            this.editor.clearSelection();
            this.script_file_query(null, null, true);
            this.config_persist();
        }.bind(this)).catch(util.handle_error.bind(this))
    },

    script_file_current : function() {
        return this.$refs.tree_script_file.selectedNodes[0];
    },


    script_file_save: function () {

        let current = this.script_file_current();
        if (current) {
            if (this.setting.editor.save_with_confirm) {
                if (!confirm("do save " + current.id  + " ?")) {
                    return;
                }
            }
        } else {
            alert("need to select one first");
            return;
        }

        this.setting.script_file.last = this.script_file.name;
        this.script_file.script = this.editor.getValue().trim();
        return axios.post("/os/file/write", {
            path : current.id,
            category: this.setting.locate.category,
            text : this.script_file.script
        }).then(function (resp) {
            util.handle_response(resp, this.console, "saved @ " + current.id);
            util.popover("#button_script_file_save", "save success", "bottom");
            return resp
        }.bind(this)).catch(util.handle_error.bind(this));

    },


    script_file_add: function(opts) {
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
                let result = QUtil.tree_locate(this.script_file_group.tree, { id : "all" },  {
                    depth_limit : 0
                });
                parent = result.target;
                parent = parent.children;
            } else {
                parent = this.script_file_group.tree;
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
            existing = QUtil.tree_locate(this.script_file_group.tree, { field_id : "id", id : node.id }, {});
        } else {
            existing = QUtil.tree_locate(parent, { field_id : "label", id : node.name }, {});
        }

        if (existing) {
            alert("节点已存在: " + existing.target.label);
            return;
        }
        parent.push(node);
        if (opts.type === "script") {
            this.script_file_save( { type : "script", name : name } );
        }
        this.script_file_save( { type : "group" } );
    },

    script_file_delete: function (opts) {
        let current = this.script_file_current();
        if (current) {
            if (!confirm("sure to delete " + current.id + " ?")) {
                return;
            }
        } else {
            alert("need to select one first");
            return;
        }

        if (!confirm("sure to delete? " + this.script_file.name)) {
            return;
        }
        axios.post("/os/file/delete", {
            path: this.script_file.name,
            category: this.setting.locate.category
        }).then(function (resp) {
            util.handle_response(resp, this.console, "script deleted @ " + this.script_file.name)
            this.script_file.name = null;
            this.script_file_list( { type : "script,group" } );
        }.bind(this)).catch(util.handle_error.bind(this));

    },

    script_file_query_prev: function (mode, carrayscript, do_not_save) {

        if (!do_not_save && this.setting.editor.save_before_run) {
            return this.script_file_save().then(function () {
                return this.script_file_query_prev(mode, carrayscript, true);
            }.bind(this));
        }

        let script = this.editor.getValue().trim();
        if (script_file.length === 0) {
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
            name : this.script_file.name,
        }, {
            timeout: this.setting.script_file.timeout || 300000
        }).then(function (resp) {
            if (resp.data.code === 404) {
                return this.script_file_query(mode, true, true);
            }
            let data = util.handle_response(resp);
            if (mode === "raw") {
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

    script_file_query: function (mode, carrayscript, do_not_save) {

        if (!do_not_save && this.setting.editor.save_before_run) {
            let script_file_save_promise = this.script_file_save();
            if (script_file_save_promise) {
                return script_file_save_promise.then(function () {
                    return this.script_file_query(mode, carrayscript, true);
                }.bind(this));
            }
            return
        }

        mode = mode  || this.setting.mode;

        this.console.text = "";
        let current = this.script_file_current();
        return axios.post("/cmd/go", {
            type : 'luafile',
            cmd : 'run',
            mode : mode,
            path : current.id,
            params : this.params
        }, {
            timeout: this.setting.script_file.timeout || 300000
        }).then(function (resp) {
            let data = util.handle_response(resp);
            if (mode === "raw") {
                let layout = [];
                if (data.error) {
                    layout.push("[error]");
                    if (typeof data.error === 'object') {
                        let error_stringify = JSON.stringify(data.error, null, 2);
                        layout.push(error_stringify);
                    } else {
                        layout.push(data.error);
                    }
                    layout.push("\n");
                }
                layout.push("[data]");
                if (typeof data.data === 'object') {
                    let data_stringify = JSON.stringify(data.data, null, 2);
                    layout.push(data_stringify);
                } else {
                    layout.push(data.data);
                }
                layout.push("\n");

                if (data.stdout) {
                    layout.push("[stdout]");
                    layout.push(data.stdout);
                    layout.push("\n");
                }

                if (data.consume) {
                    layout.push("[consume]");
                    layout.push(data.consume);
                    layout.push("\n");
                }

                this.console.text = layout.join("\n");
                return data;
            } else {
                if (data) {
                    data.refresh_view = true;
                    return this.stock_get_data_by_code(data);
                }
            }
        }.bind(this));
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
                tree_src : this.script_file_group.tree,
                tree_des : this.script_file_group.tree
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
                tree_src : this.script_file_group.tree,
                tree_des : this.script_file_group.tree
            };
        }

        $('#div_script_setting').modal('toggle');
    },

    script_file_group_tree_src_select : function(node, id) {
        this.script_setting_opts.tree_src_node = node;
    },

    script_file_group_tree_des_select : function(node, id) {
        this.script_setting_opts.tree_des_node = node;
    },

    script_file_group_move : function() {
        let tree_script_src = this.$refs.tree_script_src;
        let tree_script_des = this.$refs.tree_script_des;

        let nodes_src = tree_script_src.selectedNodes;
        let nodes_des = tree_script_des.selectedNodes;

        if (nodes_src.length === 0 || nodes_des.length === 0) {
            alert("需要选择來源及目标");
            return;
        }
    },

    script_file_group_copy : function() {

    }

};