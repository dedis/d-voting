import { ID } from 'types/configuration';
import { Results } from 'types/electionInfo';
import { ENDPOINT_EVOTING_GET_ELECTION } from './Endpoints';

const useGetResults = () => {
  async function getResults(
    electionID: ID,
    setError: React.Dispatch<any>,
    setResult: React.Dispatch<React.SetStateAction<Results[]>>,
    setIsResultSet: React.Dispatch<React.SetStateAction<boolean>>
  ) {
    const request: RequestInit = {
      method: 'GET',
    };

    try {
      const response = await fetch(ENDPOINT_EVOTING_GET_ELECTION(electionID), request);

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
