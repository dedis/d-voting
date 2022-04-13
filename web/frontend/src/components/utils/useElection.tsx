import useFetchCall from './useFetchCall';
import useFillElectionFields from './useFillElectionFields';
import { ENDPOINT_EVOTING_GET_ELECTION } from './Endpoints';

// Custom hook that fetches an election given its id and returns its
// different parameters

// TODO remove tokens everywhere
const useElection = (electionID, token) => {
  const request = {
    method: 'GET',
    headers: {
      Authorization: `Bearer ${token}`,
    },
  };

  const [data, loading, error] = useFetchCall(ENDPOINT_EVOTING_GET_ELECTION(electionID), request);
  const {
    electionTitle,
    configObj,
    status,
    pubKey,
    ballotSize,
    chunksPerBallot,
    result,
    setResult,
    setStatus,
    isResultSet,
    setIsResultSet,
  } = useFillElectionFields(data);
  return {
    loading,
    electionTitle,
    configObj,
    electionID,
    status,
    pubKey,
    ballotSize,
    chunksPerBallot,
    result,
    setResult,
    setStatus,
    isResultSet,
    setIsResultSet,
    error,
  };
};

export default useElection;
