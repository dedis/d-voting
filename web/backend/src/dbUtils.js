// This file provides utility functions to for managing admin users
//
//   node -e 'require("./dbUtils").addAdmin("1234")'
//   node -e 'require("./dbUtils").listUserPermissions("1234")'
//   node -e 'require("./dbUtils").removeAdmin("1234")'
//
// If your are running this script outside of this module, specify NODE_PATH=

const {SequelizeAdapter} = require("casbin-sequelize-adapter");
const {newEnforcer} = require("casbin");

async function initEnforcer() {
  const dbAdapter = await SequelizeAdapter.newAdapter({
    dialect: 'postgres',
    host: process.env.DATABASE_HOST,
    port: parseInt(process.env.DATABASE_PORT || '5432', 10),
    username: process.env.DATABASE_USERNAME,
    password: process.env.DATABASE_PASSWORD,
    database: 'casbin',
  });

  return newEnforcer('../model.conf', dbAdapter);
}


async function addAdmin(sciper) {
  const enforcer = await initEnforcer();
  const permissions = [
      [sciper, 'roles', 'add'],
      [sciper, 'roles', 'list'],
      [sciper, 'roles', 'remove'],
      [sciper, 'proxies', 'post'],
      [sciper, 'proxies', 'put'],
      [sciper, 'proxies', 'delete'],
      [sciper, 'election', 'create'],
  ]
  await enforcer.addPolicies(permissions);
  console.log("Successfully imported permissions for user!")
}


async function listUserPermissions(userID) {
  const enforcer = await initEnforcer();
  const userPermissions = await enforcer.getPermissionsForUser(userID)
  console.log(userPermissions)
}


async function removeAdmin(sciper) {
  const enforcer = await initEnforcer();
  await enforcer.deletePermissionsForUser(sciper)
}

module.exports = { addAdmin, listUserPermissions, removeAdmin };
