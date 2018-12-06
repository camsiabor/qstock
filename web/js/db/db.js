
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
    this.dbsize = this.dbsize || (10 * 1024 * 1024);
    this.db = window.openDatabase(this.name, this.version, this.name, this.dbsize, function (db) {
        console.log("[db]", "open", this.name, db );
    }.bind(this));
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


DB.prototype.createtable = function(tablename, keyname, fields) {
    if (!keyname) {
        keyname = "id";
    }
    if (!fields) {
        fields = [ 'data' ]
    }

    let table = {
        name : tablename,
        keyname : keyname,
        fields : fields
    };
    let sql = 'CREATE TABLE IF NOT EXISTS ' + tablename + ' (' + keyname + ' unique, ' + fields.join(",") + ')';
    return this.exec(sql, [], function () {
        this.tables[tablename] = table;
        console.log("[db]", this.name, tablename, "created");
    }.bind(this));
};

DB.prototype.droptable = function(tablename) {
    this.exec("DROP TABLE " + tablename, [], function() {
        this.tables[tablename] = null;
        console.log("[db]", this.name, tablename, "dropped");
    }.bind(this));
};

DB.prototype.update = function(tablename, keyname, fieldnames, data) {

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
            let updateq = [ "data=?" ];
            for(let i = 0; i < fieldnames.length; i++) {
                insertq.push("?");
                updateq.push(fieldnames[i] + "=?");
            }
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

        }, this.errcallback);
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

DB.prototype.query_by_id = function (tablename, keyname, ids, callback) {
    let qarray = [ ];
    for (let i = 0; i < ids.length; i++) {
        qarray.push("?");
    }
    qarray = qarray.join(",");

    let sql = "SELECT * FROM " + tablename + " WHERE " + keyname + " IN (" + qarray + ") ";
    return this.query(sql, ids, callback);
};