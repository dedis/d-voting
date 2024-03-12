const pollTransaction = (
  endpoint: (token: string) => string,
  data: any,
  interval: number,
  maxAttempts: number
) => {
  let attempts = 0;

  const request = {
    method: 'GET',
    headers: { 'Content-Type': 'application/json' },
  };

  const executePoll = async (resolve, reject): Promise<any> => {
    let response, result;
    try {
      attempts += 1;
      response = await fetch(endpoint(data), request);
      result = await response.json();
    } catch (e) {
      return reject(e);
    }

    if (!response.ok) {
      throw new Error(JSON.stringify(result));
    }

    data = result.Token;

    if (result.Status === 1) {
      return resolve(result);
    }

    if (result.Status === 2) {
      throw new Error('Transaction Rejected');
    }

    // Add a timeout
    if (attempts === maxAttempts) {
      throw new Error('Timeout');
    }

    setTimeout(executePoll, interval, resolve, reject);
  };

  return new Promise(executePoll);
};

export default pollTransaction;
