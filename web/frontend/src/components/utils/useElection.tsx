import useFetchCall from './useFetchCall';
import * as endpoints from './Endpoints';
import { useFillElectionInfo } from './FillElectionInfo';

// Custom hook that fetches an election given its id and returns its
// different parameters
const useElection = (electionID) => {
  const request = {
    method: 'GET',
  };

  const [data, loading, error] = useFetchCall(endpoints.election(electionID.toString()), request);
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
  } = useFillElectionInfo(data);

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
