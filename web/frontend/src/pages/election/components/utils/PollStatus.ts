import { Status } from 'types/election';
import { DKGInfo, NodeStatus } from 'types/node';

// https://gist.github.com/treyhuffine/b108ec8a771d3045ba8e4e3c30d9c496#file-poll-example-js
const pollElection = (
  endpoint: RequestInfo,
  request: RequestInit,
  validate: (status: Status) => boolean,
  interval: number
) => {
  const executePoll = async (resolve, reject) => {
    try {
      const response = await fetch(endpoint, request);
      const result = await response.json();

      if (!response.ok) {
        return reject(new Error(JSON.stringify(result)));
      }

      if (validate(result.Status)) {
        return resolve(result);
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
  interval: number
) => {
  const executePoll = async (resolve, reject) => {
    try {
      const response = await fetch(endpoint, request);
      const result: DKGInfo = await response.json();
      // If not initialized yet continue polling

      if (response.status === 404) {
        setTimeout(executePoll, interval, resolve, reject);
        return;
      }

      if (!response.ok) {
        return reject(new Error(JSON.stringify(result)));
      }

      if (validate(result.Status)) {
        return resolve(result);
      }

      // TODO: define the error code for the case when a node is already setup
      if (result.Error.Message.includes('already setup')) {
        return resolve(result);
      }

      if ((result.Status as NodeStatus) === NodeStatus.Failed) {
        return reject(new Error(JSON.stringify(result.Error.Message)));
      }

      setTimeout(executePoll, interval, resolve, reject);
    } catch (e) {
      return reject(e);
    }
  };

  return new Promise(executePoll);
};

export { pollElection, pollDKG };
