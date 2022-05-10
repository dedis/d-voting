import { NodeStatus, Status } from 'types/election';

// https://gist.github.com/treyhuffine/b108ec8a771d3045ba8e4e3c30d9c496#file-poll-example-js
const poll = (
  endpoint: RequestInfo,
  request: RequestInit,
  validate: (status: Status | NodeStatus) => boolean,
  interval: number,
  isDKGRequest: boolean
) => {
  const executePoll = async (resolve, reject) => {
    try {
      const response = await fetch(endpoint, request);
      const result = await response.json();

      if (!response.ok) {
        return reject(new Error(JSON.stringify(result)));
      } else if (validate(result.Status)) {
        return resolve(result);
      } else if (isDKGRequest && (result.Status as NodeStatus) === NodeStatus.Failed) {
        return reject(new Error(JSON.stringify(result.Error.message)));
      } else {
        setTimeout(executePoll, interval, resolve, reject);
      }
    } catch (e) {
      return reject(e);
    }
  };

  return new Promise(executePoll);
};

export { poll };
