import useFetchCall from './useFetchCall';
import useFillElectionFields from './useFillElectionFields';
import * as endpoints from './Endpoints';

// Custom hook that fetches an election given its id and returns its
// different parameters

// TODO remove tokens everywhere
const useElection = (electionID) => {
  const request = {
    method: 'GET',
  };

  const [data, loading, error] = useFetchCall(endpoints.election(electionID.toString()), request);
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
