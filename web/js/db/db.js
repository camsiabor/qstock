
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
    let promise = new Promise(function (resolve, reject) {
        this.db = window.openDatabase(this.name, this.version, this.name, this.dbsize);
        if (this.db) {
            console.log("[db]", "open", this.name, "size", this.dbsize, this.db );
            resolve();
        } else {
            reject();
        }
    }.bind(this));

    promise.then(function () {
        return this.query_raw_promise("SELECT * FROM sqlite_master WHERE type = 'table'");
    }.bind(this)).then(function (tables) {
        for(let i = 0; i < tables.length; i++) {
            let table = this.table_parse(tables[i]);
            this.tables[table.name] = table;
        }
    }.bind(this)).catch(function (err) {
        throw "[db] create database " + this.name + " error " + err;
    });
}




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
    tableinfo.fields = sql_created.substring(q_open_i + 1, q_close_i).split(",");
    tableinfo.fields_map = {};
    for(let i = 0; i < tableinfo.fields.length; i++) {
        let field = tableinfo.fields[i].trim();
        if (!field) {
            continue;
        }
        if (field.indexOf(" unique")) {
            field = field.substring(0, field.indexOf(" "));
            tableinfo.keyname = field;
        }
        tableinfo.fields[i] = field;
        tableinfo.fields_map[field] = field;
    }
    return tableinfo;
};

DB.prototype.table_get = function(tablename, callback, errcallback) {
    this.query_raw("SELECT * FROM sqlite_master WHERE type = ? AND name = ?", [ "table", tablename ], function (rows) {
        let table;
        if (rows && rows.length > 0) {
            table = rows[0];
            table = this.table_parse(table);
        }
        if (callback) {
            callback(table, rows);
        }
    }.bind(this), errcallback || this.errcallback());
};

DB.prototype.table_get_promise = function(tablename) {
    return new Promise(function (resolve, reject) {
        this.table_get(tablename, function (table) {
            resolve(table);
        }.bind(this), function (tx, err) {
            console.error("[db]", "[get_table_promie]", tx, err);
            reject(err);
        });
    }.bind(this));
};

DB.prototype.table_create = function(tablename, keyname, fields) {
    if (!keyname) {
        keyname = "id";
    }
    if (!fields) {
        fields = [ 'data' ]
    }
    let promise = this.table_get_promise(tablename);
    promise.then(function (table) {
        let need_to_drop = false;
        if (table) {
            let fields_map = {};
            fields_map[keyname + " unique"] = false;
            for(let i = 0; i < table.fields.length; i++) {
                let field = table.fields[i];
                fields_map[field] = false;
            }
            for(let field in fields_map) {
                if (!fields_map[field]) {
                    need_to_drop = true;
                    break;
                }
            }
        }
        if (need_to_drop) {
            return this.table_drop_promise(tablename);
        }
    }.bind(this)).then(function () {
        // droptable_promise then
    }).finally(function() {
        let sql = 'CREATE TABLE IF NOT EXISTS ' + tablename + ' (' + keyname + ' unique, ' + fields.join(",") + ')';
        this.exec(sql, [], function () {
            console.log("[db]", this.name, tablename, "created");
        }.bind(this));

    }.bind(this));


};

DB.prototype.table_drop = function(tablename) {
    this.exec("DROP TABLE " + tablename, [], function() {
        this.tables[tablename] = null;
        console.log("[db]", this.name, tablename, "dropped");
    }.bind(this));
};

DB.prototype.table_drop_promise = function(tablename) {
    return this.exec_promise("DROP TABLE " + tablename);
}



DB.prototype.update = function(tablename, keyname, fieldnames, data, callback, errcallback) {
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

    this.query_ids(tablename, keyname, ids, function (ids_exist) {

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

DB.prototype.update_promise = function(tablename, keyname, fieldname, data) {
  return new Promise(function (resolve, reject) {
      this.update(tablename, keyname, fieldname, data, function () {
          resolve();
      }.bind(this), function (tx, err) {
          console.error("[db]", "[update_promise]", tx, err);
          reject(err);
      });
  }.bind(this));
};

DB.prototype.delete_by_id = function(tablename, keyname, ids, callback, errcallback) {
    return this.db.transaction(function (tx) {
        let qarray = [];
        for(let i = 0; i < ids.length; i++) {
            qarray.push("?");
        }
        let sql = "DELETE FROM " + tablename + " WHERE " + keyname + " IN (" + qarray.join(",") + ")";
        tx.executeSql(sql, ids, function (tx, result) {
            if (callback) {
                callback(result, tx);
            }
        }, errcallback || this.errcallback);
    }.bind(this));
};

DB.prototype.delete_by_id_promise = function(tablename, keyname, ids) {
    return new Promise(function (resolve, reject) {
          this.delete_by_id(tablename, keyname, )
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

DB.prototype.query_ids = function (tablename, keyname, ids, callback, errcallback) {
    let qarray = [ ];
    for (let i = 0; i < ids.length; i++) {
        qarray.push("?");
    }
    qarray = qarray.join(",");

    let sql = "SELECT " + keyname + " FROM " + tablename + " WHERE " + keyname + " IN (" + qarray + ") ";
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

DB.prototype.query_ids_promise = function(tablename, keyname, ids) {
    return new Promise(function (resolve, reject) {
        this.query_ids(tablename, keyname, ids, function (ids) {
            resolve(ids);
        }, function (tx, err) {
            console.error("[db]", "[query_ids_promise]", tx, err);
            reject(err);
        });
    }.bind(this));
};




DB.prototype.query_by_id = function (tablename, keyname, ids, callback) {
    let qarray = [ ];
    for (let i = 0; i < ids.length; i++) {
        qarray.push("?");
    }
    qarray = qarray.join(",");

    let sql = "SELECT * FROM " + tablename + " WHERE " + keyname + " IN (" + qarray + ") ";
    return this.query(sql, ids, callback);
};

DB.prototype.query_by_id_promise = function (tablename, keyname, ids) {
    let qarray = [ ];
    for (let i = 0; i < ids.length; i++) {
        qarray.push("?");
    }
    qarray = qarray.join(",");

    let sql = "SELECT * FROM " + tablename + " WHERE " + keyname + " IN (" + qarray + ") "
    return this.query_promise(sql, ids);
}