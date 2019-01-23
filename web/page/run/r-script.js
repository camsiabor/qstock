const script_methods = {

    script_list: function (opts) {
        let path = opts.path || ".";
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
                this.script_group.tree = data;
            }
        }.bind(this)).catch(util.handle_error.bind(this));
    },

    script_select: function (noderaw, id, node) {

        if (!node) { /* active select */
            if (noderaw) {
                node = this.$refs.tree_script.getNode(noderaw.id);
            } else {
                node = this.$refs.tree_script.selectedNodes[0];
            }
            if (node) {
                this.$refs.tree_script.select(node);
            }
            return;
        }


        this.setting.script.last = this.script.name = node.id;
        return axios.post("/os/file/text", {
            path: node.id
        }).then(function (resp) {
            let text = util.handle_response(resp);
            this.script.script = text;
            this.editor.setValue(this.script.script);
            this.editor.clearSelection();
            this.script_query();
            this.config_persist();
        }.bind(this)).catch(util.handle_error.bind(this))
    },

    script_current : function() {
        return this.$refs.tree_script.selectedNodes[0];
    },


    script_save: function () {

        let current = this.script_current();
        if (current) {
            if (!confirm("do save " + current.id  + " ?")) {
                return;
            }
        } else {
            alert("need to select one first");
            return;
        }

        this.setting.script.last = this.script.name = name;
        this.script.script = this.editor.getValue().trim();
        return axios.post("/os/file/text", {
            path : current.id,
            text : this.script.script
        }).then(function (resp) {
            util.handle_response(resp, this.console, "saved @ " + current.id);
            util.popover("#button_script_save", "save success", "bottom");
            return resp
        }.bind(this)).catch(util.handle_error.bind(this));

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
        let current = this.script_current();
        if (current) {
            if (!confirm("sure to delete " + current.id + " ?")) {
                return;
            }
        } else {
            alert("need to select one first");
            return;
        }

        if (!confirm("sure to delete? " + this.script.name)) {
            return;
        }
        axios.post("/os/file/delete", {
            name: this.script.name
        }).then(function (resp) {
            util.handle_response(resp, this.console, "script deleted @ " + this.script.name)
            this.script.name = null;
            this.script_list( { type : "script,group" } );
        }.bind(this)).catch(util.handle_error.bind(this));

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