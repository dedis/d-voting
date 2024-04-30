import express from 'express';

import { addPolicy, addListPolicy, initEnforcer, isAuthorized, PERMISSIONS } from '../authManager';

export const usersRouter = express.Router();

initEnforcer().catch((e) => console.error(`Couldn't initialize enforcerer: ${e}`));

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

// This call (only for admins) allows an admin to add a role to a voter.
usersRouter.post('/add_role', (req, res, next) => {
  if (!isAuthorized(req.session.userId, PERMISSIONS.SUBJECTS.ROLES, PERMISSIONS.ACTIONS.ADD)) {
    res.status(400).send('Unauthorized - only admins allowed');
    return;
  }

  if (req.body.permission === 'vote') {
    if (!isAuthorized(req.session.userId, req.body.subject, PERMISSIONS.ACTIONS.OWN)) {
      res.status(400).send('Unauthorized - not owner of form');
    }
  }

  if ('userId' in req.body) {
    try {
      addPolicy(req.body.userId, req.body.subject, req.body.permission);
    } catch (error) {
      res.status(400).send(`Error while adding single user to roles: ${error}`);
      return;
    }
    res.set(200).send();
    next();
  } else if ('userIds' in req.body) {
    try {
      addListPolicy(req.body.userIds, req.body.subject, req.body.permission);
    } catch (error) {
      res.status(400).send(`Error while adding multiple users to roles: ${error}`);
      return;
    }
    res.set(200).send();
    next();
  } else {
    res
      .status(400)
      .send(`Error: at least one of 'userId' or 'userIds' must be send in the request`);
  }
});

// This call (only for admins) allow an admin to remove a role to a user.
usersRouter.post('/remove_role', (req, res, next) => {
  if (!isAuthorized(req.session.userId, PERMISSIONS.SUBJECTS.ROLES, PERMISSIONS.ACTIONS.REMOVE)) {
    res.status(400).send('Unauthorized - only admins allowed');
    return;
  }
  next();
});
