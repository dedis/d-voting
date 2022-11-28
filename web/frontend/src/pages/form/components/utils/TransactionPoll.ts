const pollTransaction = (
  endpoint: RequestInfo,
  data: any,
  interval: number,
  maxAttempts: number
) => {
  let attempts = 0;

  const request = {
    method: 'GET',
    headers: { 'Content-Type': 'application/json' },
    data: data,
  };

  const executePoll = async (resolve, reject) => {
    try {
      attempts += 1;
      console.log('Request:' + JSON.stringify(request));
      const response = await fetch(endpoint, request);
      const result = await response.json();

      if (!response.ok) {
        throw new Error(JSON.stringify(result));
      }

      request.data = result;

      if (result.Status === '1') {
        return resolve(result);
      }

      if (result.Status === '2') {
        throw new Error('Transaction Rejected');
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

export default pollTransaction;
