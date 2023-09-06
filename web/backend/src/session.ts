import createMemoryStore from 'memorystore';
import session from 'express-session';

const MemoryStore = createMemoryStore(session);

export const sessionStore = new MemoryStore({
  checkPeriod: 86400000, // prune expired entries every 24h
});

// Keeps an in-memory mapping between a SCIPER (userId) and its opened session
// IDs. Needed to invalidate the sessions of a user when its role changes. The
// value is a set of sessions IDs.
export const sciper2sess = new Map<number, Set<string>>();
