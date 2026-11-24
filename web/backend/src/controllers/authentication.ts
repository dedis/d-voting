import express from 'express';
import * as oauth from 'oauth4webapi';
import { sciper2sess } from '../session';
import { initEnforcer, getUserPermissions, readSCIPER, setMapAuthorization } from '../authManager';

export const authenticationRouter = express.Router();

initEnforcer().catch((e) => console.error(`Couldn't initialize enforcerer: ${e}`));

// Microsoft Entra ID authentication

// set up authentication
const tenantId = process.env.MS_ENTRA_TENANT_ID || '';
const clientId = process.env.MS_ENTRA_CLIENT_ID || '';
const redirectUri = process.env.MS_ENTRA_REDIRECT_URI || '';
const clientSecret = process.env.MS_ENTRA_CLIENT_SECRET || '';
const devLoginEnabled = process.env.REACT_APP_DEV_LOGIN === 'true';
const devLoginPSK = process.env.REACT_APP_DEV_PSK || '';

if (!devLoginEnabled && !(tenantId && clientId && redirectUri && clientSecret)) {
  throw new Error('required Microsoft Entra ID environment variables are not set');
}

if (devLoginEnabled && !devLoginPSK) {
  throw new Error('missing required dev login pre shared key. Please set env var REACT_APP_DEV_PSK');
}

const codeChallengeMethod = 'S256';

let as: oauth.AuthorizationServer;
let codeVerifier: string;
let nonce: string;

export async function initOAuth() {
  if (devLoginEnabled) {
    console.log('Microsoft Entra ID disabled because development login is enabled');
    return;
  }

  const issuer = new URL(`https://login.microsoftonline.com/${tenantId}/v2.0`);
  as = await oauth.processDiscoveryResponse(issuer, await oauth.discoveryRequest(issuer));
  console.log('Discovered authorization server');
}

// Register Microsoft Entra ID routes only when development login is disabled.
if (!devLoginEnabled) {
  const client: oauth.Client = { client_id: clientId };
  const clientAuth = oauth.ClientSecretPost(clientSecret);

  // authorization endpoint
  authenticationRouter.get('/auth-redirect', async (req, res) => {
    try {
      codeVerifier = oauth.generateRandomCodeVerifier();
      const codeChallenge = await oauth.calculatePKCECodeChallenge(codeVerifier);
      if (!as?.authorization_endpoint) {
        throw new Error('Invalid authorization endpoint');
      }
      const authorizationUrl = new URL(as.authorization_endpoint);
      authorizationUrl.searchParams.set('client_id', client.client_id);
      authorizationUrl.searchParams.set('redirect_uri', redirectUri);
      authorizationUrl.searchParams.set('response_type', 'code');
      authorizationUrl.searchParams.set('scope', 'openid email');
      authorizationUrl.searchParams.set('code_challenge', codeChallenge);
      authorizationUrl.searchParams.set('code_challenge_method', codeChallengeMethod);

      // backwards compatibility
      // https://github.com/panva/oauth4webapi/blob/222d1cc7b8e5f81ec1bbaab8ff364209e9dd7d98/examples/oidc.ts#L48
      if (as.code_challenge_methods_supported?.includes(codeChallengeMethod) !== true) {
        nonce = oauth.generateRandomNonce();
        authorizationUrl.searchParams.set('nonce', nonce);
      }

      // redirect user to Microsoft Entra ID authentication
      res.json({ url: authorizationUrl.href });
    } catch (error) {
      res.status(500).send('Failed to redirect to Microsoft Entra ID for authentication');
      console.error(error);
    }
  });

  // redirection URI
  authenticationRouter.get(redirectUri.split('/api')[1], async (req, res) => {
    try {
      const params = oauth.validateAuthResponse(
        as,
        { client_id: clientId },
        new URL(`${req.protocol}://${req.get('host')}${req.originalUrl}`)
      );
      const response = await oauth.authorizationCodeGrantRequest(
        as,
        { client_id: clientId },
        clientAuth,
        params,
        redirectUri,
        codeVerifier
      );
      const result = await oauth.processAuthorizationCodeResponse(as, client, response, {
        expectedNonce: nonce,
        requireIdToken: true,
      });
      const claims = oauth.getValidatedIdTokenClaims(result);
      if (!(claims?.uniqueid && claims?.family_name && claims?.given_name)) {
        throw new Error('Invalid authentication response');
      }
      req.session.userId = parseInt(claims.uniqueid as string, 10);
      req.session.lastName = claims.family_name as string;
      req.session.firstName = claims.given_name as string;

      // log user in into React app
      const sciperSessions = sciper2sess.get(req.session.userId) || new Set<string>();
      sciperSessions.add(req.sessionID);
      sciper2sess.set(req.session.userId, sciperSessions);

      console.log(`user ${req.session.userId} successfully logged in`);

      res.redirect('/logged');
    } catch (error) {
      res.status(500).send('Failed to log in user');
      console.error(error);
    }
  });
}

authenticationRouter.get('/get_dev_login/:userId', (req, res) => {
  if (!devLoginEnabled) {
    const err = `/get_dev_login can only be called with REACT_APP_DEV_LOGIN===true: ${process.env.REACT_APP_DEV_LOGIN}`;
    console.error(err);
    res.status(500).send(err);
    return;
  }
  if (req.params.userId === undefined) {
    const err = 'no userId given';
    console.error(err);
    res.status(400).send(err);
    return;
  }
  if (req.params.psk === undefined || req.params.psk != devLoginPSK) {
    const err = 'incorrect or absent dev login psk';
    console.error(err);
    res.status(403).send(err);
    return;
  }
  try {
    req.session.userId = readSCIPER(req.params.userId);
    req.session.firstName = 'sciper-#';
    req.session.lastName = req.params.userId;
  } catch (e) {
    const err = `Invalid userId: ${e}`;
    console.error(err);
    res.status(500).send(err);
    return;
  }

  const sciperSessions = sciper2sess.get(req.session.userId) || new Set<string>();
  sciperSessions.add(req.sessionID);
  sciper2sess.set(req.session.userId, sciperSessions);

  res.redirect('/logged');
});

// This endpoint serves to log out from the app by clearing the session.
authenticationRouter.post('/logout', (req, res) => {
  if (req.session.userId === undefined) {
    res.status(400).send('not logged in');
    return;
  }

  const { userId } = req.session;

  req.session.destroy(() => {
    const a = sciper2sess.get(userId as number);
    if (a !== undefined) {
      a.delete(req.sessionID);
      sciper2sess.set(userId as number, a);
    }
    res.redirect('/');
  });
});

// As the user is logged on the app via this express but must also
// be logged into react. This endpoint serves to send to the client (actually to react)
// the information of the current user.
authenticationRouter.get('/personal_info', async (req, res) => {
  if (!req.session.userId) {
    res.status(401).send('Unauthenticated');
    return;
  }
  const userPermissions = await getUserPermissions(req.session.userId);
  res.set('Access-Control-Allow-Origin', '*');
  res.json({
    sciper: req.session.userId,
    lastName: req.session.lastName,
    firstName: req.session.firstName,
    isLoggedIn: true,
    authorization: Object.fromEntries(setMapAuthorization(userPermissions)),
  });
});
