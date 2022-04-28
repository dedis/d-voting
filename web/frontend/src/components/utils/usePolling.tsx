import { useEffect, useState } from 'react';
import { ID } from 'types/configuration';
import { STATUS } from 'types/election';
import * as endpoints from './Endpoints';

// https://gist.github.com/treyhuffine/b108ec8a771d3045ba8e4e3c30d9c496#file-poll-example-js
const usePoll = (
  endpoint: RequestInfo,
  request: RequestInit,
  validate: (status: STATUS) => boolean,
  interval: number,
  setError: (error: any) => void
) => {
  console.log('Start poll...');

  const executePoll = async (resolve) => {
    console.log('- poll');
    try {
      const response = await fetch(endpoint, request);

      if (!response.ok) {
        const js = await response.json();
        throw new Error(JSON.stringify(js));
      } else {
        let result = await response.json();
        if (validate(result.Status)) {
          return resolve(result);
        } else {
          setTimeout(executePoll, interval, resolve);
        }
      }
    } catch (e) {
      setError(e);
    }
  };

  return new Promise(executePoll);
};

const usePollElectionStatus = (electionID: ID, statusToMatch: STATUS) => {
  const [data, setData] = useState(null);
  const [error, setError] = useState(null);

  const interval = 1000;
  const request = {
    method: 'GET',
  };

  const match = (status: STATUS) => status === statusToMatch;

  usePoll(endpoints.election(electionID), request, match, interval, setError).then((dataReceived) =>
    setData(dataReceived)
  );

  return [data, error];
};

const usePollDKGStatus = (electionID: ID, statusToMatch: STATUS) => {
  const [data, setData] = useState(null);
  const [error, setError] = useState(null);

  // Might be interesting to set larger interval depending on the status to poll
  const interval = 1000;
  const request = {
    method: 'GET',
  };

  const match = (status: STATUS) => status === statusToMatch;

  usePoll(endpoints.dkgActor(electionID), request, match, interval, setError).then((dataReceived) =>
    setData(dataReceived)
  );

  return [data, error];
};

// handler
/*const simulateServerRequestTime = (interval) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve();
    }, interval);
  });
};

const TIME_FOR_AUTH_PROVIDER = 10000;
const TIME_TO_CREATE_NEW_USER = 2000;

let fakeUser = null;
const createUser = (() => {
  setTimeout(() => {
    fakeUser = {
      id: '123',
      username: 'testuser',
      name: 'Test User',
      createdAt: Date.now(),
    };
  }, TIME_FOR_AUTH_PROVIDER + TIME_TO_CREATE_NEW_USER);
})();

const mockApi = async () => {
  await simulateServerRequestTime(500);
  return fakeUser;
};

const validateUser = (user) => !!user;
const POLL_INTERVAL = 1000;
const pollForNewUser = poll({
  fn: mockApi,
  validate: validateUser,
  interval: POLL_INTERVAL,
})
  .then((user) => console.log(user))
  .catch((err) => console.error(err));*/
