import express from 'express';
import axios, { AxiosError, Method } from 'axios';
import cookieParser from 'cookie-parser';
import session from 'express-session';
import morgan from 'morgan';
import kyber from '@dedis/kyber';
import crypto from 'crypto';
import lmdb, { RangeOptions } from 'lmdb';
import xss from 'xss';
import createMemoryStore from 'memorystore';
import { Enforcer, newEnforcer } from 'casbin';

const MemoryStore = createMemoryStore(session);
const SUBJECT_ROLES = 'roles';
const SUBJECT_PROXIES = 'proxies';
const SUBJECT_ELECTION = 'election';

const ACTION_LIST = 'list';
const ACTION_ACTION_REMOVE = 'remove';
const ACTION_ADD = 'add';
const ACTION_PUT = 'put';
const ACTION_POST = 'post';
const ACTION_DELETE = 'delete';
const ACTION_CREATE = 'create';
// store is used to store the session
const store = new MemoryStore({
  checkPeriod: 86400000, // prune expired entries every 24h
});

// Keeps an in-memory mapping between a SCIPER (userid) and its opened session
// IDs. Needed to invalidate the sessions of a user when its role changes. The
// value is a set of sessions IDs.
const sciper2sess = new Map<number, Set<string>>();

const app = express();

app.use(morgan('tiny'));

let enf: Enforcer;

const enforcerLoading = newEnforcer('model.conf', 'policy.csv');
const port = process.env.PORT || 5000;

Promise.all([enforcerLoading])
  .then((res) => {
    [enf] = res;
    console.log(`🛡 Casbin loaded`);
    app.listen(port);
    console.log(`🚀 App is listening on port ${port}`);
  })
  .catch((err) => {
    console.error('❌ failed to start:', err);
  });

function isAuthorized(sciper: number | undefined, subject: string, action: string): boolean {
  return enf.enforceSync(sciper, subject, action);
}

declare module 'express-session' {
  export interface SessionData {
    userid: number;
    firstname: string;
    lastname: string;
    role: string;
  }
}

// Express-session
app.set('trust-proxy', 1);

app.use(cookieParser());
const oneDay = 1000 * 60 * 60 * 24;
app.use(
  session({
    secret: process.env.SESSION_SECRET as string,
    saveUninitialized: true,
    cookie: { maxAge: oneDay },
    resave: false,
    store: store,
  })
);

app.use(express.json());
app.use(express.urlencoded({ extended: true }));

/*
 * Access control
 * */
// app.use((req, res, next) => {
//   const begin = req.url.split('?')[0];
//   let role = 'everyone';
//   if (req.session.userid && req.session.role) {
//     role = req.session.role;
//   }

//   if (accessConfig[role].includes(begin)) {
//     next();
//   } else {
//     res.status(400).send('Unauthorized');
//   }
// });

// This endpoint allows anyone to get a "default" proxy. Clients can still use
// the proxy of their choice thought.

app.get('/api/config/proxy', (req, res) => {
  res.status(200).send(process.env.DELA_NODE_URL);
});

const usersDB = lmdb.open<'admin' | 'operator', number>({
  path: `${process.env.DB_PATH}dvoting-users`,
});

// This is via this endpoint that the client request the tequila key, this key
// will then be used for redirection on the tequila server
app.get('/api/get_teq_key', (req, res) => {
  const body = `urlaccess=${process.env.FRONT_END_URL}/api/control_key\nservice=Evoting\nrequest=name,firstname,email,uniqueid,allunits`;
  axios
    .post('https://tequila.epfl.ch/cgi-bin/tequila/createrequest', body)
    .then((response) => {
      const key = response.data.split('\n')[0].split('=')[1];
      const url = `https://tequila.epfl.ch/cgi-bin/tequila/requestauth?requestkey=${key}`;
      res.json({ url: url });
    })
    .catch((error: AxiosError) => {
      console.log('message:', error.message);
      res.status(500).send(`failed to request Tequila authentication: ${error.message}`);
    });
});

// Here the client will send the key he/she received from the tequila, it is
// then verified on the tequila. If the key is valid, the user is then logged
// in the website through this backend
app.get('/api/control_key', (req, res) => {
  const userKey = req.query.key;
  const body = `key=${userKey}`;

  axios
    .post('https://tequila.epfl.ch/cgi-bin/tequila/fetchattributes', body)
    .then((resa) => {
      if (!resa.data.includes('status=ok')) {
        throw new Error('Login did not work');
      }

      const sciper = resa.data.split('uniqueid=')[1].split('\n')[0];
      const lastname = resa.data.split('\nname=')[1].split('\n')[0];
      const firstname = resa.data.split('\nfirstname=')[1].split('\n')[0];

      const role = usersDB.get(sciper) || '';

      
      req.session.userid = parseInt(sciper, 10);
      req.session.lastname = lastname;
      req.session.firstname = firstname;
      req.session.role = role;

      const a = sciper2sess.get(req.session.userid) || new Set<string>();
      a.add(req.sessionID);
      sciper2sess.set(sciper, a);

      res.redirect('/logged');
    })
    .catch((error) => {
      res.status(500).send('Login did not work');
      console.log(error);
    });
});

// This endpoint serves to logout from the app by clearing the session.
app.post('/api/logout', (req, res) => {
  if (req.session.userid === undefined) {
    res.status(400).send('not logged in');
  }

  const { userid } = req.session;

  req.session.destroy(() => {
    const a = sciper2sess.get(userid as number);
    if (a !== undefined) {
      a.delete(req.sessionID);
      sciper2sess.set(userid as number, a);
    }
    res.redirect('/');
  });
});

// This function helps us convert the double list of the authorization
// returned by the casbin function getFilteredPolicy to a map that link
// an object to the action authorized
// list[0] contains the policies so list[i][0] is the sciper
// list[i][1] is the subject and list[i][2] is the action
function setMapAuthorization(list: string[][]): Map<String, Array<String>> {
  const m = new Map<String, Array<String>>();
  for (let i = 0; i < list.length; i += 1) {
    const subject = list[i][1];
    const action = list[i][2];
    if (m.has(subject)) {
      m.get(subject)?.push(action);
    } else {
      m.set(subject, [action]);
    }
  }
  console.log(m);
  return m;
}

// As the user is logged on the app via this express but must also be logged in
// the react. This endpoint serves to send to the client (actually to react)
// the information of the current user.
app.get('/api/personal_info', (req, res) => {
  enf.getFilteredPolicy(0, String(req.session.userid)).then((list) => {
    res.set('Access-Control-Allow-Origin', '*');
    if (req.session.userid) {
      res.json({
        sciper: req.session.userid,
        lastname: req.session.lastname,
        firstname: req.session.firstname,
        role: req.session.role,
        islogged: true,
        authorization: Object.fromEntries(setMapAuthorization(list)),
      });
    } else {
      res.json({
        sciper: 0,
        lastname: '',
        firstname: '',
        role: '',
        islogged: false,
        authorization: {},
      });
    }
  });
});

// ---
// Users role
// ---
// This call allow a user that is admin to get the list of the people that have
// a special role (not a voter).
app.get('/api/user_rights', (req, res) => {
  if (!isAuthorized(req.session.userid, SUBJECT_ROLES, ACTION_LIST)) {
    res.status(400).send('Unauthorized - only admins allowed');
    return;
  }
  const opts: RangeOptions = {};
  const users = Array.from(
    usersDB.getRange(opts).map(({ key, value }) => ({ id: '0', sciper: key, role: value }))
  );
  res.json(users);
});

// This call (only for admins) allow an admin to add a role to a voter.
app.post('/api/add_role', (req, res) => {
  if (!isAuthorized(req.session.userid, SUBJECT_ROLES, ACTION_ADD)) {
    res.status(400).send('Unauthorized - only admins allowed');
    return;
  }

  const { sciper } = req.body;
  const { role } = req.body;

  usersDB.put(sciper, role).catch((error) => {
    res.status(500).send('Failed to add role');
    console.log(error);
  });
});

// This call (only for admins) allow an admin to remove a role to a user.
app.post('/api/remove_role', (req, res) => {
  if (!isAuthorized(req.session.userid, SUBJECT_ROLES, ACTION_ACTION_REMOVE)) {
    res.status(400).send('Unauthorized - only admins allowed');
    return;
  }

  const { sciper } = req.body;
  usersDB
    .remove(sciper)
    .then(() => {
      const sessionIDs = sciper2sess.get(sciper);
      if (sessionIDs !== undefined) {
        sessionIDs.forEach((_, sessionID) => {
          store.destroy(sessionID);
        });
      }

      res.status(200).send('Removed');
    })
    .catch((error) => {
      res.status(500).send('Remove role failed');
      console.log(error);
    });
});

// ---
// end of users role
// ---

// ---
// Proxies
// ---
const proxiesDB = lmdb.open<string, string>({ path: `${process.env.DB_PATH}proxies` });
app.post('/api/proxies', (req, res) => {
  if (!isAuthorized(req.session.userid, SUBJECT_PROXIES, ACTION_POST)) {
    res.status(400).send('Unauthorized - only admins and operators allowed');
    return;
  }
  try {
    const bodydata = req.body;
    proxiesDB.put(bodydata.NodeAddr, bodydata.Proxy);
    console.log('put', bodydata.NodeAddr, '=>', bodydata.Proxy);
    res.status(200).send('ok');
  } catch (error: any) {
    res.status(500).send(error.toString());
  }
});

app.put('/api/proxies/:nodeAddr', (req, res) => {
  if (!isAuthorized(req.session.userid, SUBJECT_PROXIES, ACTION_PUT)) {
    res.status(400).send('Unauthorized - only admins and operators allowed');
    return;
  }

  let { nodeAddr } = req.params;

  nodeAddr = decodeURIComponent(nodeAddr);

  const proxy = proxiesDB.get(nodeAddr);

  if (proxy === undefined) {
    res.status(404).send('not found');
    return;
  }
  try {
    const bodydata = req.body;
    if (bodydata.Proxy === undefined) {
      res.status(400).send('bad request, proxy is undefined');
      return;
    }

    proxiesDB.put(nodeAddr, bodydata.Proxy);
    console.log('put', nodeAddr, '=>', bodydata.Proxy);
    res.status(200).send('ok');
  } catch (error: any) {
    res.status(500).send(error.toString());
  }
});

app.delete('/api/proxies/:nodeAddr', (req, res) => {
  if (!isAuthorized(req.session.userid, SUBJECT_PROXIES, ACTION_DELETE)) {
    res.status(400).send('Unauthorized - only admins and operators allowed');
    return;
  }

  let { nodeAddr } = req.params;

  nodeAddr = decodeURIComponent(nodeAddr);

  const proxy = proxiesDB.get(nodeAddr);

  if (proxy === undefined) {
    res.status(404).send('not found');
    return;
  }

  try {
    proxiesDB.remove(nodeAddr);
    console.log('remove', nodeAddr, '=>', proxy);
    res.status(200).send('ok');
  } catch (error: any) {
    res.status(500).send(error.toString());
  }
});

app.get('/api/proxies', (req, res) => {
  const output = new Map<string, string>();
  proxiesDB.getRange({}).forEach((entry) => {
    output.set(entry.key, entry.value);
  });

  res.status(200).json({ Proxies: Object.fromEntries(output) });
});

app.get('/api/proxies/:nodeAddr', (req, res) => {
  const { nodeAddr } = req.params;

  const proxy = proxiesDB.get(decodeURIComponent(nodeAddr));

  if (proxy === undefined) {
    res.status(404).send('not found');
    return;
  }

  res.status(200).json({
    NodeAddr: nodeAddr,
    Proxy: proxy,
  });
});

// ---
// end of proxies
// ---

// get payload creates a payload with a signature on it
function getPayload(dataStr: string) {
  let dataStrB64 = Buffer.from(dataStr).toString('base64url');
  while (dataStrB64.length % 4 !== 0) {
    dataStrB64 += '=';
  }

  const hash: Buffer = crypto.createHash('sha256').update(dataStrB64).digest();

  const edCurve = kyber.curve.newCurve('edwards25519');

  const priv = Buffer.from(process.env.PRIVATE_KEY as string, 'hex');
  const pub = Buffer.from(process.env.PUBLIC_KEY as string, 'hex');

  const scalar = edCurve.scalar();
  scalar.unmarshalBinary(priv);

  const point = edCurve.point();
  point.unmarshalBinary(pub);

  const sign = kyber.sign.schnorr.sign(edCurve, scalar, hash);

  const payload = {
    Payload: dataStrB64,
    Signature: sign.toString('hex'),
  };

  return payload;
}

// sendToDela signs the message and sends it to the dela proxy. It makes no
// authentication check.
function sendToDela(dataStr: string, req: express.Request, res: express.Response) {
  let payload = getPayload(dataStr);

  // we strip the `/api` part: /api/form/xxx => /form/xxx
  let uri = process.env.DELA_NODE_URL + req.baseUrl.slice(4);
  // boolean to check
  let redirectToDefaultProxy = true;
  // in case this is a DKG init request, we must also update the payload.

  const dkgInitRegex = /\/evoting\/services\/dkg\/actors$/;
  if (uri.match(dkgInitRegex)) {
    const dataStr2 = JSON.stringify({ FormID: req.body.FormID });
    payload = getPayload(dataStr2);
    redirectToDefaultProxy = false;
  }

  // in case this is a DKG setup request, we must update the payload.
  const dkgSetupRegex = /\/evoting\/services\/dkg\/actors\/.*$/;
  if (uri.match(dkgSetupRegex)) {
    const dataStr2 = JSON.stringify({ Action: req.body.Action });
    payload = getPayload(dataStr2);

    // If setup don't redirect to default proxy, if 'computePubshares' then keep
    // default proxy
    if (req.body.Action === 'setup') {
      redirectToDefaultProxy = false;
    }
  }

  // in case this is a DKG init or setup request, we must extract the proxy addr
  if (!redirectToDefaultProxy) {
    const proxy = req.body.Proxy;

    if (proxy === undefined) {
      res.status(400).send('proxy undefined in body');
      return;
    }
    uri = proxy + req.baseUrl.slice(4);
  }

  console.log('sending payload:', JSON.stringify(payload), 'to', uri);

  axios({
    method: req.method as Method,
    url: uri,
    data: payload,
    headers: {
      'Content-Type': 'application/json',
    },
  })
    .then((resp) => {
      res.status(200).send(resp.data);
    })
    .catch((error: AxiosError) => {
      let resp = '';
      if (error.response) {
        resp = JSON.stringify(error.response.data);
      }

      res
        .status(500)
        .send(`failed to proxy request: ${req.method} ${uri} - ${error.message} - ${resp}`);
    });
}

// Secure /api/evoting to admins and operators
app.use('/api/evoting/*', (req, res, next) => {
  if (!isAuthorized(req.session.userid, SUBJECT_ELECTION, ACTION_CREATE)) {
    res.status(400).send('Unauthorized - only admins and operators allowed');
    return;
  }
  next();
});

// https://stackoverflow.com/a/1349426
function makeid(length: number) {
  let result = '';
  const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  const charactersLength = characters.length;
  for (let i = 0; i < length; i += 1) {
    result += characters.charAt(Math.floor(Math.random() * charactersLength));
  }
  return result;
}

app.delete('/api/evoting/forms/:formID', (req, res) => {
  const { formID } = req.params;

  const edCurve = kyber.curve.newCurve('edwards25519');

  const priv = Buffer.from(process.env.PRIVATE_KEY as string, 'hex');
  const pub = Buffer.from(process.env.PUBLIC_KEY as string, 'hex');

  const scalar = edCurve.scalar();
  scalar.unmarshalBinary(priv);

  const point = edCurve.point();
  point.unmarshalBinary(pub);

  const sign = kyber.sign.schnorr.sign(edCurve, scalar, Buffer.from(formID));

  // we strip the `/api` part: /api/form/xxx => /form/xxx
  const uri = process.env.DELA_NODE_URL + xss(req.url.slice(4));

  axios({
    method: req.method as Method,
    url: uri,
    headers: {
      Authorization: sign.toString('hex'),
    },
  })
    .then((resp) => {
      res.status(200).send(resp.data);
    })
    .catch((error: AxiosError) => {
      let resp = '';
      if (error.response) {
        resp = JSON.stringify(error.response.data);
      }

      res
        .status(500)
        .send(`failed to proxy request: ${req.method} ${uri} - ${error.message} - ${resp}`);
    });
});

// This API call is used redirect all the calls for DELA to the DELAs nodes.
// During this process the data are processed : the user is authenticated and
// controlled. Once this is done the data are signed before the are sent to the
// DELA node To make this work, react has to redirect to this backend all the
// request that needs to go the DELA nodes
app.use('/api/evoting/*', (req, res) => {
  if (!req.session.userid) {
    res.status(400).send('Unauthorized');
    return;
  }

  const bodyData = req.body;

  // special case for voting
  const regex = /\/api\/evoting\/forms\/.*\/vote/;
  if (req.baseUrl.match(regex)) {
    // We must set the UserID to know who this ballot is associated to. This is
    // only needed to allow users to cast multiple ballots, where only the last
    // ballot is taken into account. To preserve anonymity the web-backend could
    // translate UserIDs to another random ID.
    // bodyData.UserID = req.session.userid.toString();
    bodyData.UserID = makeid(10);
  }

  const dataStr = JSON.stringify(bodyData);

  sendToDela(dataStr, req, res);
});

// Handles any requests that don't match the ones above
app.get('*', (req, res) => {
  console.log('404 not found');
  const url = new URL(req.url, `http://${req.headers.host}`);
  res.status(404).send(`not found ${xss(url.toString())}`);
});
