// This file provides utility functions to handle an lmdb database. You can use
// the node CLI to call those functions:
//
//   node -e 'require("./dbUtils").addAdmin("./dvoting-users", 1234)'
//   node -e 'require("./dbUtils").listEls("./dvoting-users")'
//   node -e 'require("./dbUtils").removeEl("./dvoting-users", 1234)'
//
// If your are running this script outside of this module, specify NODE_PATH=

const lmdb = require('lmdb');

const addAdmin = (dbPath, sciper) => {
    const usersDB = lmdb.open({ path: dbPath });

    usersDB.put(String(sciper), "admin").then(() => {
        console.log("ok");
    }).catch((error) => {
        console.log(error);
    });
}

const listEls = (dbPath) => {
    const db = lmdb.open({ path: dbPath });

    db.getRange({}).forEach(
        ({ key, value }) => { console.log(`'${key}' => '${value}' \t | \t  (${typeof key}) => (${typeof key})`) }
    )
}

const removeEl = (dbPath, key) => {
    const db = lmdb.open({ path: dbPath });

    db.remove(String(key))
        .then(() => {
            console.log(`key '${key}' removed`)
        })
        .catch((error) => {
            console.log(`error: ${error}`);
        });
}

module.exports = { addAdmin, listEls, removeEl }