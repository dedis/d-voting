#!/usr/bin/env ts-node

/*
Backend CLI, currently providing 3 commands for user management:
  npx cli addAdmin --sciper 1234
  npx cli listUserPermissions --sciper 1234
  npx cli removeAdmin --sciper 1234
*/

import { Command, InvalidArgumentError } from 'commander';
import { SequelizeAdapter } from 'casbin-sequelize-adapter';
import { newEnforcer } from 'casbin';
import { curve } from '@dedis/kyber';
import * as fs from 'fs';
import { PERMISSIONS } from './authManager';

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

// Imports a list of SCIPERS from a file to allow to vote on a specific election
// the .voters.example file is available as an example
program
  .command('addVoters')
  .description('Assigns a list of SCIPERs to an Election as Voters')
  .requiredOption('-e, --election-id <char>', 'ID of the election')
  .requiredOption('-sf, --scipers-file <char>', 'File with line-separated list of SCIPERs')
  .action(async ({ electionId, scipersFile }) => {
    fs.readFile(scipersFile, 'utf8', async (err: any, data: string) => {
      if (err) {
        throw new InvalidArgumentError(`Faced a problem trying to process your file: \n ${err}`);
      }
      const scipers: Array<string> = data.split('\n');
      const policies = [];
      for (let i = 0; i < scipers.length; i += 1) {
        const sciper: number = Number(scipers[i]);
        if (Number.isNaN(sciper)) {
          throw new InvalidArgumentError(`SCIPER '${sciper}' on line ${i + 1} is not a number - exiting!`);
        }
        if (sciper > 999999 || sciper < 100000) {
          throw new InvalidArgumentError(
            `SCIPER '${sciper}' on line ${i + 1} is outside acceptable range (100000..999999) - exiting!`
          );
        }
        policies[i] = [scipers[i], electionId, PERMISSIONS.ACTIONS.VOTE];
      }
      const enforcer = await initEnforcer();
      await enforcer.addPolicies(policies);
      console.log('Added Voting policies successfully!');
    });
  });

program.parse();
