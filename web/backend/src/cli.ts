#!/usr/bin/env ts-node

/*
Backend CLI, currently providing 3 commands for user management:
  npx cli addAdmin --sciper 1234
  npx cli listUserPermissions --sciper 1234
  npx cli removeAdmin --sciper 1234
*/

import { Command, InvalidArgumentError } from 'commander';
import { curve, Group } from '@dedis/kyber';
import * as fs from 'fs';
import request from 'request';
import ShortUniqueId from 'short-unique-id';
import { initEnforcer, PERMISSIONS, readSCIPER } from './authManager';

const program = new Command();

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
        try {
          policies[i] = [readSCIPER(scipers[i]), electionId, PERMISSIONS.ACTIONS.VOTE];
        } catch (e) {
          throw new InvalidArgumentError(
            `SCIPER '${scipers[i]}' on line ${i + 1} is not a valid sciper: ${e}`
          );
        }
      }
      const enforcer = await initEnforcer();
      await enforcer.addPolicies(policies);
      console.log('Added Voting policies successfully!');
    });
  });

function getRequest(url: string): Promise<{ response: request.Response; body: any }> {
  return new Promise((resolve, reject) => {
    request.get({ url: url, followRedirect: false }, (err, response, body) => {
      if (err) {
        reject(err);
      } else {
        resolve({ response, body });
      }
    });
  });
}

function postRequest(
  url: string,
  cookie: string,
  data: any
): Promise<{ response: request.Response; body: any }> {
  return new Promise((resolve, reject) => {
    request.post(
      {
        url: url,
        headers: {
          Cookie: cookie,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
      },
      (err, response, body) => {
        if (err) {
          reject(err);
        } else {
          resolve({ response, body });
        }
      }
    );
  });
}

function encodeBallot(
  formSelectId: string,
  choices: number[],
  ballotSize: number,
  chunksPerBallot: number
): string[] {
  let encodedBallot = `select:${Buffer.from(formSelectId).toString('base64')}:${choices.join(
    ','
  )}\n\n`;

  const ebSize = Buffer.byteLength(encodedBallot);

  if (ebSize < ballotSize) {
    const padding = new ShortUniqueId({ length: ballotSize - ebSize });
    encodedBallot += padding();
  }

  const chunkSize = 29;
  const ballotChunks: string[] = [];

  // divide into chunksPerBallot chunks, where 1 character === 1 byte
  for (let i = 0; i < chunksPerBallot; i += 1) {
    const start = i * chunkSize;
    // substring(start, start + chunkSize), if (start + chunkSize) > string.length
    // then (start + chunkSize) is treated as if it was equal to string.length
    ballotChunks.push(encodedBallot.substring(start, start + chunkSize));
  }

  return ballotChunks;
}

export function encryptVote(vote: string, dkgKey: Buffer, edCurve: Group) {
  // embed the vote into a curve point
  const M = edCurve.point().embed(Buffer.from(vote));
  // dkg public key as a point on the EC
  const keyBuff = dkgKey;
  const p = edCurve.point();
  p.unmarshalBinary(keyBuff); // unmarshal dkg public key
  const pubKeyPoint = p.clone(); // get the point corresponding to the dkg public key

  const k = edCurve.scalar().pick(); // ephemeral private key
  const K = edCurve.point().mul(k); // ephemeral DH public key

  const S = edCurve.point().mul(k, pubKeyPoint); // ephemeral DH shared secret
  const C = S.add(S, M); // message blinded with secret

  // (K,C) are what we'll send to the backend
  return [K.marshalBinary(), C.marshalBinary()];
}

program
  .command('vote')
  .description('Votes multiple times - only works with REACT_APP_DEV_LOGIN=true')
  .requiredOption('-f, --frontend <char>', 'URL of frontend')
  .requiredOption('-e, --election-id <char>', 'ID of the election')
  .requiredOption('-b, --ballots <number>', 'how many ballots to cast')
  .requiredOption('-a, --admin <char>', 'admin ID')
  .action(async ({ frontend, electionId, ballots, admin }) => {
    console.log(`Going to cast ${ballots} ballots in election-id ${electionId} over ${frontend}`);

    console.log(`Getting proxies`);
    const responseProxies = await getRequest(`${frontend}/api/proxies`);
    const proxies = JSON.parse(responseProxies.body).Proxies;
    const proxy = Object.values(proxies)[0];

    console.log(`Getting data of form ${electionId}`);
    const responseForm = await getRequest(`${proxy}/evoting/forms/${electionId}`);
    console.log(responseForm.body);
    const form = JSON.parse(responseForm.body);
    const formPubkey = form.Pubkey;
    if (!('Selects' in form.Configuration.Scaffold[0])) {
      throw new Error('Only support forms with 1 scaffold of type selects');
    }
    const formSelect = form.Configuration.Scaffold[0].Selects[0];
    const formSelectId = formSelect.ID;
    if (formSelect.MinN !== 1) {
      throw new Error('Only forms with MinN === 1 supported');
    }

    // Always vote for the first choice.
    // ballotsize, chunksperballot
    const choices1 = formSelect.Choices.map(() => 0);
    choices1[0] = 1;
    const ballotChunks1 = encodeBallot(
      formSelectId,
      choices1,
      form.BallotSize,
      form.ChunksPerBallot
    );
    if (ballotChunks1.length !== 1) {
      throw new Error('Should get exactly one ballot-chunk');
    }
    const EGPair1 = encryptVote(
      ballotChunks1[0],
      Buffer.from(formPubkey, 'hex'),
      curve.newCurve('edwards25519')
    );
    const choices2 = formSelect.Choices.map(() => 0);
    choices2[1] = 1;
    const ballotChunks2 = encodeBallot(
      formSelectId,
      choices2,
      form.BallotSize,
      form.ChunksPerBallot
    );
    const EGPair2 = encryptVote(
      ballotChunks2[0],
      Buffer.from(formPubkey, 'hex'),
      curve.newCurve('edwards25519')
    );

    console.log('Getting login cookie');
    const { response } = await getRequest(`${frontend}/api/get_dev_login/${admin}`);
    if (response.headers['set-cookie']?.length !== 1) {
      throw new Error("Didn't get cookie");
    }
    const loginCookie = response.headers['set-cookie']![0];

    console.log('Casting ballots');
    for (let i = 0; i < ballots; i += 1) {
      const start = Date.now();
      // Have 1/3 vote for choice 1, 2/3 for choice 2
      const EGPair = i % 3 === 0 ? EGPair1 : EGPair2;
      // eslint-disable-next-line no-await-in-loop
      const responseCast = await postRequest(
        `${frontend}/api/evoting/forms/${electionId}/vote`,
        loginCookie,
        { Ballot: [{ K: Array.from(EGPair[0]), C: Array.from(EGPair[1]) }], UserId: `${admin}` }
      );
      if (responseCast.response.statusCode !== 200) {
        console.log(responseCast.response.headers);
      }
      if (i % 10 === 0) {
        console.log(`Casting ballot ${i} took ${Date.now() - start}ms`);
      }
      // await new Promise(resolve => setTimeout(resolve, 1000));
    }
  });

program.parse();
