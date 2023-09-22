import express from 'express';
import morgan from 'morgan';
import xss from 'xss';
import { authenticationRouter } from './controllers/authentication';
import { usersRouter } from './controllers/users';
import { proxiesRouter } from './controllers/proxies';
import { delaRouter } from './controllers/dela';
import { setupSession } from './session';

const app = express();

app.use(morgan('tiny'));

declare module 'express-session' {
  // This overrides express-session
  export interface SessionData {
    userId: number;
    firstName: string;
    lastName: string;
  }
}

app.set('trust-proxy', 1);
setupSession(app);

app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// This endpoint allows anyone to get a "default" proxy. Clients can still use
// the proxy of their choice thought.
app.get('/api/config/proxy', (req, res) => {
  res.status(200).send(process.env.DELA_NODE_URL);
});

app.use('/api', authenticationRouter);
app.use('/api', usersRouter);
app.use('/api/proxies', proxiesRouter);
app.use('/api/evoting', delaRouter);

// Handles any requests that don't match the ones above
app.get('*', (req, res) => {
  console.log('404 not found');
  const url = new URL(req.url, `http://${req.headers.host}`);
  res.status(404).send(`not found ${xss(url.toString())}`);
});

const serveOnPort = process.env.BACKEND_PORT || 5000;
app.listen(serveOnPort);
console.log(`🚀 App is listening on port ${serveOnPort}`);
