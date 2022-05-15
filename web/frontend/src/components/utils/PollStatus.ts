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
      } else if (validate(result.Status)) {
        return resolve(result);
      } else {
        setTimeout(executePoll, interval, resolve, reject);
      }
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

      if (!response.ok) {
        // If not initialized yet continue polling
        if (response.status == 404) {
          setTimeout(executePoll, interval, resolve, reject);
        } else {
          return reject(new Error(JSON.stringify(result)));
        }
      } else if (validate(result.Status)) {
        return resolve(result);
      } else if ((result.Status as NodeStatus) === NodeStatus.Failed) {
        return reject(new Error(JSON.stringify(result.Error.Message)));
      } else {
        setTimeout(executePoll, interval, resolve, reject);
      }
    } catch (e) {
      return reject(new Error(JSON.stringify(e.message)));
    }
  };

  return new Promise(executePoll);
};

export { pollElection, pollDKG };