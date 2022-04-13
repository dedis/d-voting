import useFetchCall from './useFetchCall';
import useFillElectionFields from './useFillElectionFields';
import { ENDPOINT_EVOTING_GET_ELECTION } from './Endpoints';
import { ID } from 'types/configuration';

// Custom hook that fetches an election given its id and returns its
// different parameters

const useElection = (electionID: ID) => {
  const request: RequestInit = {
    method: 'GET',
  };
  const endpoint = `${ENDPOINT_EVOTING_GET_ELECTION}/${electionID}`;
  console.log(endpoint);
  const [data, loading, error] = useFetchCall(endpoint, request);
  const {
    status,
    setStatus,
    pubKey,
    result,
    setResult,
    chunksPerBallot,
    ballotSize,
    configObj,
    isResultSet,
    setIsResultSet,
  } = useFillElectionFields(data);

  return {
    loading,
    electionID,
    status,
    setStatus,
    pubKey,
    result,
    setResult,
    chunksPerBallot,
    ballotSize,
    configObj,
    isResultSet,
    setIsResultSet,
    error,
  };
};

export default useElection;
