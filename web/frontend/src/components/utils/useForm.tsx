import useFetchCall from './useFetchCall';
import * as endpoints from './Endpoints';
import { useFillFormInfo } from './FillFormInfo';
import { ID } from 'types/configuration';
import { useContext } from 'react';
import { ProxyContext } from 'index';

// Custom hook that fetches a form given its id and returns its
// different parameters
const useForm = (formID: ID) => {
  const pctx = useContext(ProxyContext);

  const request = {
    method: 'GET',
  };
  const [data, loading, error] = useFetchCall(endpoints.form(pctx.getProxy(), formID), request);
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
    voters,
  } = useFillFormInfo(data);

  return {
    loading,
    formID,
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
    voters,
    error,
  };
};

export default useForm;
