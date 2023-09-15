import createMemoryStore from 'memorystore';
import session from 'express-session';
import cookieParser from 'cookie-parser';
import { Express } from 'express';

const MemoryStore = createMemoryStore(session);

export const sessionStore = new MemoryStore({
  checkPeriod: 86400000, // prune expired entries every 24h
});

// Keeps an in-memory mapping between a SCIPER (userId) and its opened session
// IDs. Needed to invalidate the sessions of a user when its role changes. The
// value is a set of sessions IDs.
export const sciper2sess = new Map<number, Set<string>>();

export const setupSession = (app: Express) => {
  app.use(cookieParser());
  const oneDay = 1000 * 60 * 60 * 24;
  app.use(
    session({
      secret: process.env.SESSION_SECRET as string,
      saveUninitialized: true,
      cookie: { maxAge: oneDay },
      resave: false,
      store: sessionStore,
    })
  );
};
