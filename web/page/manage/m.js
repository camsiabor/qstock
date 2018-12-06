
let vue = new Vue({
    el: '#dcontainer',
    /* [data] ------------------------------------------------------------------- */
    data: {
        cmd : {
            type : "redis",
            types : [ "redis", "js", "lua", "os" ],
            content : "",
            template : {
                "redis" : {
                    "db" : "def",
                    "cmd" : "HGETALL",
                    "args" : [ "000001" ]
                },
                "os" : {
                    "cmd" : "ping",
                    "args" : [ "127.0.0.1", "-n", "10" ],
                    "timeout" : 5
                }
            }
        },
        console : {
            text : ""
        }
    },
    /* [methods] ------------------------------------------------------------------- */
    methods : {
        config_persist : function() {
        },

        config_load : function() {
        },

        cmd_select : function(cmdtype) {
            let mode;
            this.cmd.type = cmdtype;
            switch(this.cmd.type) {
                case "redis" :
                    mode = "json";
                    break;
                case "js" :
                    mode = "json";
                    break;
                case "os" :
                    mode = "json";
                    break;
            }
            this.editor.session.setMode("ace/mode/" + mode);
            let template = this.cmd.template[cmdtype];
            if (template) {
                let templatestr = JSON.stringify(template, null, 2);
                this.editor.setValue(templatestr);
            }
            this.editor.clearSelection();
        },

        cmd_go : function() {
            let cmd = this.editor.getValue().trim();
            cmd = JSON.parse(cmd);
            cmd.type = this.cmd.type;
            axios.post("/cmd/go", cmd).then(function(resp) {
                let data = bscommon.handle_response(resp, this.console);
                switch( cmd.type) {
                    case "redis":
                        break;
                    case "lua":
                        break;
                    case "js":
                        break;
                    case "os":
                        let consoletext = [];
                        if (data.timeout) {
                            consoletext.push("@timeout");
                            consoletext.push(data.timeout);
                            consoletext.push("");
                        }
                        if (data.stdout) {
                            consoletext.push("@stdout");
                            consoletext.push(data.stdout);
                            consoletext.push("");
                        }
                        if (data.stderr) {
                            consoletext.push("@stderr");
                            consoletext.push(data.stderr);
                            consoletext.push("");
                        }
                        this.console.text = consoletext.join("\n");
                        break;
                }

            }.bind(this)).catch(bscommon.handle_error.bind(this));
        },

        /* [init] ------------------------------------------------------------------- */

        init_editor : function() {
            this.editor =  ace.edit("editor", {
                mode: "ace/mode/lua",
                selectionStyle: "text",
                highlightActiveLine: true,
                highlightSelectedWord: true,
                cursorStyle: "ace",
                newLineMode: "unix",
                fontSize: "0.8em"
            });

            this.editor.setOption("wrap", "free");
            this.editor.setTheme("ace/theme/github");
        }
    },

    /* [mount] ------------------------------------------------------------------- */
    mounted : function() {
        this.config_load();
        this.init_editor();
        this.cmd_select(this.cmd.type);
    }
});
