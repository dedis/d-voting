import { Status } from 'types/form';
import { DKGInfo, NodeStatus } from 'types/node';

// https://gist.github.com/treyhuffine/b108ec8a771d3045ba8e4e3c30d9c496#file-poll-example-js
const pollForm = (
  endpoint: RequestInfo,
  request: RequestInit,
  validate: (status: Status) => boolean,
  interval: number,
  maxAttempts: number
) => {
  let attempts = 0;

  const executePoll = async (resolve, reject) => {
    try {
      attempts += 1;
      const response = await fetch(endpoint, request);
      const result = await response.json();

      if (!response.ok) {
        throw new Error(JSON.stringify(result));
      }

      if (validate(result.Status)) {
        return resolve(result);
      }

      // Add a timeout
      if (attempts === maxAttempts) {
        throw new Error('Timeout');
      }

      setTimeout(executePoll, interval, resolve, reject);
    } catch (e) {
      return reject(e);
    }
  };

  return new Promise(executePoll);
};

const pollDKG = (
  endpoint: RequestInfo,
  request: RequestInit,
  validate: (status: NodeStatus) => boolean,
  interval: number,
  maxAttempts: number,
  status: (status: NodeStatus) => void
): Promise<DKGInfo> => {
  let attempts = 0;

  const executePoll = async (resolve, reject) => {
    try {
      attempts += 1;
      const response = await fetch(endpoint, request);
      const result: DKGInfo = await response.json();

      // Add a timeout
      if (attempts === maxAttempts) {
        throw new Error('Timeout');
      }
      // If not initialized yet continue polling
      if (response.status === 404) {
        setTimeout(executePoll, interval, resolve, reject);
        return;
      }

      if (!response.ok) {
        throw new Error(JSON.stringify(result));
      }

      status(result.Status);

      if (validate(result.Status)) {
        return resolve(result);
      }

      // TODO: define the error code for the case when a node is already setup
      // Ignore error to "allow" the setup of a node multiple times
      if (result.Error.Message.includes('setup() was already called, only one call is allowed')) {
        return resolve(result);
      }

      // TODO: define the error code for the case when a node is already initialized
      // Ignore error to "allow" the initialization of a node multiple times
      if (result.Error.Message.includes('actor already exists for formID')) {
        return resolve(result);
      }

      if ((result.Status as NodeStatus) === NodeStatus.Failed) {
        throw new Error(JSON.stringify(result.Error.Message));
      }

      setTimeout(executePoll, interval, resolve, reject);
    } catch (e) {
      return reject(e);
    }
  };

  return new Promise<DKGInfo>(executePoll);
};

export { pollForm, pollDKG };
