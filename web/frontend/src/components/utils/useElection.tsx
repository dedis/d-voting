import useFetchCall from './useFetchCall';
import * as endpoints from './Endpoints';
import { useFillElectionInfo } from './FillElectionInfo';
import { ID } from 'types/configuration';

// Custom hook that fetches an election given its id and returns its
// different parameters
const useElection = (electionID: ID) => {
  const request = {
    method: 'GET',
  };
  const [data, loading, error] = useFetchCall(endpoints.election(electionID), request);
  const {
    status,
    setStatus,
    pubKey,
    roster,
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
    roster,
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
