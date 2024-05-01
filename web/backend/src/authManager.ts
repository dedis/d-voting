import { SequelizeAdapter } from 'casbin-sequelize-adapter';
import { Enforcer, newEnforcer } from 'casbin';

export const PERMISSIONS = {
  SUBJECTS: {
    ROLES: 'roles',
    PROXIES: 'proxies',
    ELECTION: 'election',
  },
  ACTIONS: {
    LIST: 'list',
    REMOVE: 'remove',
    ADD: 'add',
    PUT: 'put',
    POST: 'post',
    DELETE: 'delete',
    OWN: 'own',
    CREATE: 'create',
    VOTE: 'vote',
  },
};

let authEnforcer: Enforcer;

/*
We use the postgres adapter to store the Casbin policies
we initialize the adapter with the connection string and the migrate option
the connection string has the following format:
postgres://username:password@host:port/database
the migrate option is used to create the tables if they don't exist, we set it to false because we create the tables manually
*/
export async function initEnforcer(): Promise<Enforcer> {
  if (authEnforcer === undefined) {
    const dbAdapter = await SequelizeAdapter.newAdapter({
      dialect: 'postgres',
      host: process.env.DATABASE_HOST,
      port: parseInt(process.env.DATABASE_PORT || '5432', 10),
      username: process.env.DATABASE_USERNAME,
      password: process.env.DATABASE_PASSWORD,
      database: 'casbin',
    });
    authEnforcer = await newEnforcer('src/model.conf', dbAdapter);
  }
  return authEnforcer;
}

export function isAuthorized(sciper: number | undefined, subject: string, action: string): boolean {
  return authEnforcer.enforceSync(sciper, subject, action);
}

export async function getUserPermissions(userID: number) {
  return authEnforcer.getFilteredPolicy(0, String(userID));
}

export async function addPolicy(userID: string, subject: string, permission: string) {
  await authEnforcer.addPolicy(userID, subject, permission);
  await authEnforcer.loadPolicy();
}

export async function addListPolicy(userIDs: string[], subject: string, permission: string) {
  const promises = userIDs.map((userID) => authEnforcer.addPolicy(userID, subject, permission));
  try {
    await Promise.all(promises);
  } catch (error) {
    // At least one policy update has failed, but we need to reload ACLs anyway for the succeeding ones
    await authEnforcer.loadPolicy();
    throw new Error(`Failed to add policies for all users: ${error}`);
  }
}

export async function assignUserPermissionToOwnElection(userID: string, ElectionID: string) {
  return authEnforcer.addPolicy(userID, ElectionID, PERMISSIONS.ACTIONS.OWN);
}

export async function revokeUserPermissionToOwnElection(userID: string, ElectionID: string) {
  return authEnforcer.removePolicy(userID, ElectionID, PERMISSIONS.ACTIONS.OWN);
}

// This function helps us convert the double list of the authorization
// returned by the casbin function getFilteredPolicy to a map that link
// an object to the action authorized
// list[0] contains the policies so list[i][0] is the sciper
// list[i][1] is the subject and list[i][2] is the action
export function setMapAuthorization(list: string[][]): Map<String, Array<String>> {
  const userRights = new Map<String, Array<String>>();
  for (let i = 0; i < list.length; i += 1) {
    const subject = list[i][1];
    const action = list[i][2];
    if (userRights.has(subject)) {
      userRights.get(subject)?.push(action);
    } else {
      userRights.set(subject, [action]);
    }
  }
  return userRights;
}

// Reads a SCIPER from a string and returns the number. If the SCIPER is not in
// the range between 100000 and 999999, an error is thrown.
export function readSCIPER(s: string): number {
  const n = parseInt(s, 10);
  if (isNaN(n)) {
    throw new Error(`${s} is not a number`);
  }
  if (n < 100000 || n > 999999) {
    throw new Error(`SCIPER is out of range. ${n} is not between 100000 and 999999`);
  }
  return n;
}
