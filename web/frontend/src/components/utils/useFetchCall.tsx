import { useEffect, useState } from 'react';

// Custom hook to fetch data from an endpoint
const useFetchCall = (endpoint: RequestInfo, request: RequestInit) => {
  const [data, setData] = useState(null);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const response = await fetch(endpoint, request);
        if (!response.ok) {
          setError(new Error(await response.text()));
        } else {
          let dataReceived = await response.json();
          setData(dataReceived);
          setLoading(false);
        }
      } catch (e) {
        setError(e);
      }
    };

    fetchData().catch(console.error);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [endpoint]);

  return [data, loading, error];
};

export default useFetchCall;
