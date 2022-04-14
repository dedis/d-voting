import useFetchCall from './useFetchCall';
import { ENDPOINT_EVOTING_GET_ELECTION } from './Endpoints';
import { ID } from 'types/configuration';
import { useFillElectionInfo } from './FillElectionInfo';

// Custom hook that fetches an election given its id and returns its
// different parameters

const useElection = (electionID: ID) => {
  const request: RequestInit = {
    method: 'GET',
  };
  const [data, loading, error] = useFetchCall(ENDPOINT_EVOTING_GET_ELECTION(electionID), request);
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
