import express from 'express';
import axios, { AxiosError } from 'axios';
import { sciper2sess } from '../session';
import { getUserPermissions, setMapAuthorization } from '../authManager';

export const authenticationRouter = express.Router();

// This is via this endpoint that the client request the tequila key, this key
// will then be used for redirection on the tequila server
authenticationRouter.get('/get_teq_key', (req, res) => {
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
authenticationRouter.get('/control_key', (req, res) => {
  const userKey = req.query.key;
  const body = `key=${userKey}`;

  axios
    .post('https://tequila.epfl.ch/cgi-bin/tequila/fetchattributes', body)
    .then((response) => {
      if (!response.data.includes('status=ok')) {
        throw new Error('Login did not work');
      }

      const sciper = response.data.split('uniqueid=')[1].split('\n')[0];
      const lastname = response.data.split('\nname=')[1].split('\n')[0];
      const firstname = response.data.split('\nfirstname=')[1].split('\n')[0];

      req.session.userId = parseInt(sciper, 10);
      req.session.lastName = lastname;
      req.session.firstName = firstname;

      const sciperSessions = sciper2sess.get(req.session.userId) || new Set<string>();
      sciperSessions.add(req.sessionID);
      sciper2sess.set(sciper, sciperSessions);

      res.redirect('/logged');
    })
    .catch((error) => {
      res.status(500).send('Login did not work');
      console.log(error);
    });
});

// This endpoint serves to log out from the app by clearing the session.
authenticationRouter.post('/logout', (req, res) => {
  if (req.session.userId === undefined) {
    res.status(400).send('not logged in');
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
