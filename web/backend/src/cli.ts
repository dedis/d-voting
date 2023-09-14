#!/usr/bin/env ts-node

/*
Backend CLI, currently providing 3 commands for user management:
  npx cli addAdmin --sciper 1234
  npx cli listUserPermissions --sciper 1234
  npx cli removeAdmin --sciper 1234
*/

import { Command } from 'commander';
import { SequelizeAdapter } from 'casbin-sequelize-adapter';
import { newEnforcer } from 'casbin';
import { curve } from '@dedis/kyber';

const program = new Command();

async function initEnforcer() {
  const dbAdapter = await SequelizeAdapter.newAdapter({
    dialect: 'postgres',
    host: process.env.DATABASE_HOST,
    port: parseInt(process.env.DATABASE_PORT || '5432', 10),
    username: process.env.DATABASE_USERNAME,
    password: process.env.DATABASE_PASSWORD,
    database: 'casbin',
  });

  return newEnforcer('src/model.conf', dbAdapter);
}

program
  .command('addAdmin')
  .description('Given a SCIPER number, the owner would gain full admin permissions')
  .requiredOption('-s, --sciper <char>', 'user SCIPER')
  .action(async ({ sciper }) => {
    const enforcer = await initEnforcer();
    const permissions = [
      [sciper, 'roles', 'add'],
      [sciper, 'roles', 'list'],
      [sciper, 'roles', 'remove'],
      [sciper, 'proxies', 'post'],
      [sciper, 'proxies', 'put'],
      [sciper, 'proxies', 'delete'],
      [sciper, 'election', 'create'],
    ];
    await enforcer.addPolicies(permissions);
    console.log('Successfully imported permissions for user!');
  });

program
  .command('listUserPermissions')
  .description('Lists the permissions -if any- of the owner of a given SCIPER')
  .requiredOption('-s, --sciper <char>', 'user SCIPER')
  .action(async ({ sciper }) => {
    const enforcer = await initEnforcer();
    const userPermissions = await enforcer.getPermissionsForUser(sciper);
    console.log(userPermissions);
  });

program
  .command('removeAdmin')
  .description('Given a SCIPER number, the owner would lose all admin privileges -if any-')
  .requiredOption('-s, --sciper <char>', 'user SCIPER')
  .action(async ({ sciper }) => {
    const enforcer = await initEnforcer();
    await enforcer.deletePermissionsForUser(sciper);
    console.log('Permissions removed successfully!');
  });

program
  .command('keygen')
  .description('Create a new keypair for the .env')
  .action(() => {
    const ed25519 = curve.newCurve('edwards25519');
    const priv = ed25519.scalar().pick();
    const pub = ed25519.point().mul(priv);
    console.log('Please store the following keypair in your configuration file:');
    console.log(`PRIVATE_KEY=${priv}`);
    console.log(`PUBLIC_KEY=${pub}`);
  });

program.parse();
