import { ENDPOINT_EVOTING_RESULT } from './Endpoints';

const useGetResults = () => {
  async function getResults(electionID, token, setError, setResult, setIsResultSet) {
    const request = {
      method: 'POST',
      body: JSON.stringify({ ElectionID: electionID, Token: token }),
    };
    try {
      const response = await fetch(ENDPOINT_EVOTING_RESULT, request);

      if (!response.ok) {
        throw Error(response.statusText);
      } else {
        let data = await response.json();
        setResult(data.Result);
        setIsResultSet(true);
      }
    } catch (error) {
      setError(error);
    }
  }
  return { getResults };
};

export default useGetResults;
