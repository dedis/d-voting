import express from 'express';

import { isAuthorized, PERMISSIONS } from '../authManager';

export const usersRouter = express.Router();

// This call allows a user that is admin to get the list of the people that have
// a special role (not a voter).
usersRouter.get('/user_rights', (req, res) => {
  if (!isAuthorized(req.session.userId, PERMISSIONS.SUBJECTS.ROLES, PERMISSIONS.ACTIONS.LIST)) {
    res.status(400).send('Unauthorized - only admins allowed');
    return;
  }
  const users: {
    id: string;
    sciper: number;
    role: 'admin' | 'operator';
  }[] = [];
  res.json(users);
});

// This call (only for admins) allow an admin to add a role to a voter.
usersRouter.post('/add_role', (req, res, next) => {
  if (!isAuthorized(req.session.userId, PERMISSIONS.SUBJECTS.ROLES, PERMISSIONS.ACTIONS.ADD)) {
    res.status(400).send('Unauthorized - only admins allowed');
    return;
  }

  const { sciper } = req.body;

  // The sciper has to contain 6 numbers
  if (sciper > 999999 || sciper < 100000) {
    res.status(400).send('Sciper length is incorrect');
    return;
  }
  next();
  // Call https://search-api.epfl.ch/api/ldap?q=228271, if the answer is
  // empty then sciper unknown, otherwise add it in userDB
});

// This call (only for admins) allow an admin to remove a role to a user.

usersRouter.post('/remove_role', (req, res, next) => {
  if (!isAuthorized(req.session.userId, PERMISSIONS.SUBJECTS.ROLES, PERMISSIONS.ACTIONS.REMOVE)) {
    res.status(400).send('Unauthorized - only admins allowed');
    return;
  }
  next();
});
