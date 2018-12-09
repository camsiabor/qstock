
function DB(opt) {
    for(let k in opt) {
        this[k] = opt[k];
    }
    if (!this.name) {
        throw "need to specify database name";
    }
    this.tables = {};
    this.mode = "simple";
    this.desc = this.desc || this.name;
    this.version = this.version || "1";
    this.dbsize = this.dbsize || (32 * 1024 * 1024);
    this.db = window.openDatabase(this.name, this.version, this.name, this.dbsize);
    if (this.db) {
        console.log("[db]", "open", this.name, "size", this.dbsize, this.db );
    } else {
        throw "[db] create database " + this.name + " error ";
    }

    let promises = [];
    let schema = this.schema;
    if (schema) {
        for(let tablename in schema) {
            let profile = schema[tablename];
            if (!profile) {
                continue;
            }
            let promise = this.table_create_promise(tablename, profile.keyname, profile.fields, profile);
            promises.push(promise);
        }
    }

    let promise = this.table_get_promise();
    promises.push(promise);

    let me = this;
    Promise.all(promises).then(function (r) {
        opt.callback && opt.callback("success", r, me, me.db);
    }).catch(function (err) {
        opt.callback && opt.callback("error", err, me, me.db);
    });

}


DB.new_db_promise = function(opt) {
    return new Promise(function (resolve, reject) {
        opt.callback = function (msg, r, db) {
            if (msg === 'error') {
                reject(r);
                return;
            }
            resolve(db);
        };
        try {
            new DB(opt);
        } catch (e) {
            console.error("[db]", "[new_db_promise]", e);
            reject(e);
        }
    });
};

DB.prototype.errcallback = function(tx, err) {
    console.error(tx, err);
};

DB.prototype.exec = function(sql, args, callback, errcallback) {
    return this.db.transaction(function (tx) {
        args = args || [];
        tx.executeSql(sql, args, callback, errcallback || this.errcallback);
    }.bind(this))
};

DB.prototype.exec_promise = function(sql, args) {
    args = args || [];
    let p = new Promise(function (resolve, reject) {
        this.db.transaction(function (tx) {
            tx.executeSql(sql, args, function (tx, result) {
                resolve(result.rows);
            });
        }, function (tx, err) {
            console.error("[db]", "[exec_promise]", tx, err);
            reject(err);
        });
    }.bind(this));
    return p;
};

DB.prototype.get_keyname = function(tablename) {
    let table = this.tables[tablename];
    if (!table) {
        this.table_get(tablename);
        throw "table not found " + tablename;
    }
    return table.keyname;
};


DB.prototype.table_parse = function(tableinfo) {
    let sql_created = tableinfo.sql;
    let q_open_i = sql_created.indexOf('(');
    let q_close_i = sql_created.indexOf(")");
    let fields = sql_created.substring(q_open_i + 1, q_close_i).split(",");
    tableinfo.fields = [];
    tableinfo.fields_map = {};
    for(let i = 0; i < fields.length; i++) {
        let field = fields[i].trim();
        if (!field) {
            continue;
        }
        if (field.indexOf(" unique") > 0) {
            field = field.substring(0, field.indexOf(" "));
            tableinfo.keyname = field;
        }
        tableinfo.fields[i] = field;
        tableinfo.fields_map[field] = field;
    }
    return tableinfo;
};

DB.prototype.args_flatten_qs = function(args) {
    let qs = [];
    for(let i = 0; i < args.length; i++) {
        qs.push("?");
    }
    return qs.join(",");
};

DB.prototype.table_get = function(tablenames, callback, errcallback) {
    let sql = "SELECT * FROM sqlite_master WHERE type = 'table' ";
    if (tablenames) {
        if (typeof tablenames === 'string') {
            tablenames = [ tablenames ];
        }
        let qs = this.args_flatten_qs(tablenames);
        sql = sql + " AND name in (" + qs + ") ";
    } else {
        tablenames = [];
    }
    this.query_raw(sql, tablenames, function (rows) {
        let tables = [];
        if (rows && rows.length > 0) {
            for(let i = 0; i < rows.length; i++) {
                let table = this.table_parse(rows[i]);
                this.tables[table.name] = table;
                tables.push(table);
            }
        }
        if (callback) {
            callback(tables);
        }
    }.bind(this), errcallback || this.errcallback());
};

DB.prototype.table_get_promise = function(tablename) {
    return new Promise(function (resolve, reject) {
        this.table_get(tablename, function (tables) {
            resolve(tables);
        }.bind(this), function (tx, err) {
            console.error("[db]", "[get_table_promie]", tx, err);
            reject(err);
        });
    }.bind(this));
};

DB.prototype.table_create = function(tablename, keyname, fields, options, callback, errcallback) {
    if (!keyname) {
        keyname = "id";
    }
    if (!fields) {
        fields = [ 'data' ]
    }
    let need_to_drop = false;

    let promise = this.table_get_promise(tablename);

    promise.then(function (tables) {
        let table;
        if (tables) {
            table = tables[0];
        }
        if (table) {
            if (table.keyname === keyname) {
                let fields_map = {};
                for(let i = 0; i < fields.length; i++) {
                    let field = fields[i];
                    fields_map[field] = false;
                }
                for(let i = 0; i < table.fields.length; i++) {
                    let field = table.fields[i];
                    fields_map[field] = true;
                }
                for(let field in fields_map) {
                    if (!fields_map[field]) {
                        need_to_drop = true;
                        break;
                    }
                }
            } else {
                need_to_drop = true;
            }
        }
        if (need_to_drop) {
            return this.table_drop_promise(tablename);
        }
    }.bind(this))
    .then(function () {
        let sql = 'CREATE TABLE IF NOT EXISTS ' + tablename + ' (' + keyname + ' unique, ' + fields.join(",") + ')';
        return this.exec_promise(sql, [])
    }.bind(this))
    .then(function () {
        console.log("[db]", this.name, tablename, "created");
        if (callback) {
            callback(tablename);
        }
    });

    return promise;

};

DB.prototype.table_create_promise = function(tablename, keyname, fields, options) {
    return new Promise(function (resolve, reject) {
        this.table_create(tablename, keyname, fields, options, function (r) {
            resolve(r)
        }.bind(this), function (err) {
            console.error("[db]", "[table_create_promise]", err);
            reject(err);
        })
    }.bind(this));
};

DB.prototype.table_drop = function(tablename, callback) {
    this.exec("DROP TABLE " + tablename, [], function() {
        this.tables[tablename] = null;
        console.log("[db]", this.name, tablename, "dropped");
        if (callback) {
            callback(tablename);
        }
    }.bind(this));
};

DB.prototype.table_drop_promise = function(tablename) {
    return this.exec_promise("DROP TABLE " + tablename);
};



DB.prototype.update = function(tablename, fieldnames, data, callback, errcallback) {
    let keyname = this.get_keyname(tablename);
    if (!(data instanceof Array)) {
        data = [ data ];
    }
    let ids = [];
    for(let i = 0; i < data.length; i++) {
        let id = data[i][keyname];
        if (id) {
            ids.push(id);
        }
    }

    fieldnames = fieldnames || [];

    this.query_ids(tablename, ids, function (ids_exist) {

        let ids_exist_map = {};
        for(let i = 0; i < ids_exist.length; i++) {
            let id = ids_exist[i];
            ids_exist_map[id] = true;
        }
        this.db.transaction(function (tx) {
            let insertq = [ "?","?" ];
            let updateq = [  ];
            for(let i = 0; i < fieldnames.length; i++) {
                insertq.push("?");
                updateq.push(fieldnames[i] + "=?");
            }
            updateq.push("data=?");
            let fieldnamesex = fieldnames.concat([ "data", keyname ]);
            let sql_insert = "INSERT INTO " + tablename + " (" + fieldnamesex.join(",") +  ") VALUES (" + insertq.join(",") + ")";
            let sql_update = "UPDATE " + tablename + " SET " + updateq.join(",") + " WHERE " + keyname + " = ?";

            if (!(data instanceof Array)) {
                data = [ data ];
            }
            for(let i = 0; i < data.length; i++) {
                let one = data[i];
                let str = JSON.stringify(one, null, 0);
                let id = one[keyname];
                let args = [];
                for(let f = 0; f < fieldnames.length; f++) {
                    let field = fieldnames[f];
                    let fval = one[field];
                    if (typeof fval === 'undefined') {
                        fval = "";
                    }
                    args.push(fval);
                }
                args.push(str);
                args.push(id);
                if (id) {
                    let sql = ids_exist_map[id] ? sql_update : sql_insert;
                    tx.executeSql(sql, args);
                }
            }
            callback && callback();
        }, errcallback || this.errcallback);
    }.bind(this));
};

DB.prototype.update_promise = function(tablename, fieldname, data) {
    return new Promise(function (resolve, reject) {
        this.update(tablename, fieldname, data, function () {
            resolve();
        }.bind(this), function (tx, err) {
            console.error("[db]", "[update_promise]", tx, err);
            reject(err);
        });
    }.bind(this));
};

DB.prototype.delete_by_id = function(tablename, ids, callback, errcallback) {
    let keyname = this.get_keyname(tablename);
    return this.db.transaction(function (tx) {
        let qs = this.args_flatten_qs(ids);
        let sql = "DELETE FROM " + tablename + " WHERE " + keyname + " IN (" + qs + ")";
        tx.executeSql(sql, ids, function (tx, result) {
            if (callback) {
                callback(result, tx);
            }
        }, errcallback || this.errcallback);
    }.bind(this));
};

DB.prototype.delete_by_id_promise = function(tablename, ids) {
    return new Promise(function (resolve, reject) {
          this.delete_by_id(tablename, id, function (r) {
              resolve(r);
          }.bind(this), function (tx, err) {
              console.error("[db]", "[delete_by_id_promise]", tx, err);
              reject(err);
          }.bind(this));
    }.bind(this));
};

DB.prototype.delete_all = function(tablename, callback, errcallback) {
    return this.db.transaction(function (tx) {
        let sql = "DELETE from " + tablename;
        tx.executeSql(sql, [], function (tx, result) {
            if (callback) {
                callback(result, tx);
            }
            console.info("[db]", this.name, tablename, "clear");
        }.bind(this));
    }.bind(this), errcallback || this.errcallback);
};



DB.prototype.query_raw = function(sql, args, callback, errcallback) {
    return this.db.transaction(function (tx) {
        tx.executeSql(sql, args, function (tx, result) {
            callback(result.rows, tx);
        }, errcallback || this.errcallback);
    }.bind(this));
};

DB.prototype.query_raw_promise = function(sql, args) {
    return this.exec_promise(sql, args);
};

DB.prototype.query = function (sql, args, callback, errcallback) {
    return this.db.transaction(function (tx) {
        tx.executeSql(sql, args, function (tx, result) {
            let resultarray = [];
            let rows = result.rows;
            let len = rows.length;
            for(let i = 0; i < len; i++) {
               let row = rows.item(i);
               let str = row['data'];
               if (str) {
                   let one = JSON.parse(str);
                   resultarray.push(one);
               }
            }
            callback(resultarray, tx);
        }, errcallback || this.errcallback);
    }.bind(this));
};

DB.prototype.query_promise = function (sql, args) {
    let p = new Promise(function (resolve, reject) {
        this.db.transaction(function (tx) {
            tx.executeSql(sql, args, function (tx, result) {
                let resultarray = [];
                let rows = result.rows;
                let len = rows.length;
                for(let i = 0; i < len; i++) {
                    let row = rows.item(i);
                    let str = row['data'];
                    if (str) {
                        let one = JSON.parse(str);
                        resultarray.push(one);
                    }
                }
                resolve(resultarray);
            }, function (tx, err) {
                console.error(tx, err);
                reject(err);
            });
        });
    }.bind(this));
    return p;
};

DB.prototype.query_ids = function (tablename, ids, callback, errcallback) {
    let keyname = this.get_keyname(tablename);
    let qs = this.args_flatten_qs(ids);
    let sql = "SELECT " + keyname + " FROM " + tablename + " WHERE " + keyname + " IN (" + qs + ") ";
    return this.db.transaction(function (tx) {
        tx.executeSql(sql, ids, function (tx, result) {
            let resultarray = [];
            let rows = result.rows;
            let len = rows.length;
            for(let i = 0; i < len; i++) {
                let row = rows.item(i);
                let id = row[keyname];
                resultarray.push(id);
            }
            callback(resultarray, tx);
        }, errcallback || this.errcallback);
    }.bind(this));
};

DB.prototype.query_ids_promise = function(tablename, ids) {
    return new Promise(function (resolve, reject) {
        this.query_ids(tablename, ids, function (ids) {
            resolve(ids);
        }, function (tx, err) {
            console.error("[db]", "[query_ids_promise]", tx, err);
            reject(err);
        });
    }.bind(this));
};

DB.prototype.query_by_id = function (tablename, ids, callback) {
    let keyname = this.get_keyname(tablename);
    let qs = this.args_flatten_qs(ids);
    let sql = "SELECT * FROM " + tablename + " WHERE " + keyname + " IN (" + qs + ") ";
    return this.query(sql, ids, callback);
};

DB.prototype.query_by_id_promise = function (tablename, ids) {
    let keyname = this.get_keyname(tablename);
    let qs = this.args_flatten_qs(ids);
    let sql = "SELECT * FROM " + tablename + " WHERE " + keyname + " IN (" + qs + ") ";
    return this.query_promise(sql, ids);
}