import { STATUS } from 'types/election';

// https://gist.github.com/treyhuffine/b108ec8a771d3045ba8e4e3c30d9c496#file-poll-example-js
const poll = (
  endpoint: RequestInfo,
  request: RequestInit,
  validate: (status: STATUS) => boolean,
  interval: number
) => {
  console.log('Start poll...');

  const executePoll = async (resolve, reject) => {
    console.log('- poll');

    const response = await fetch(endpoint, request);
    const result = await response.json();

    if (!response.ok) {
      return reject(new Error(JSON.stringify(result)));
    } else if (validate(result.Status)) {
      return resolve(result);
    } else {
      setTimeout(executePoll, interval, resolve, reject);
    }
  };

  return new Promise(executePoll);
};

export { poll };
