import { ID } from 'types/configuration';
import { Results } from 'types/election';
import * as endpoints from 'components/utils/Endpoints';
import { useContext } from 'react';
import { ProxyContext } from 'index';

const useGetResults = () => {
  const pctx = useContext(ProxyContext);

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
      const response = await fetch(endpoints.election(pctx.getProxy(), electionID), request);

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
