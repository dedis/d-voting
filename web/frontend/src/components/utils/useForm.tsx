import { fetchCall } from './fetchCall';
import * as endpoints from './Endpoints';
import { useFillFormInfo } from './FillFormInfo';
import { ID } from 'types/configuration';
import { useContext, useEffect, useState } from 'react';
import { ProxyContext } from 'index';

// Custom hook that fetches a form given its id and returns its
// different parameters
const useForm = (formID: ID) => {
  const pctx = useContext(ProxyContext);
  const [loading, setLoading] = useState(true);
  const [data, setData] = useState(null);
  const [error, setError] = useState(null);

  useEffect(() => {
    fetchCall(
      endpoints.form(pctx.getProxy(), formID),
      {
        method: 'GET',
      },
      setData,
      setLoading
    ).catch(setError);
  }, [pctx, formID]);
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
