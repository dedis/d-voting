import { useEffect, useState } from 'react';
import { ID } from 'types/configuration';

import { FormInfo, LightFormInfo, Results, Status } from 'types/form';

const useFillFormInfo = (formData: FormInfo) => {
  const [id, setId] = useState<ID>('');
  const [status, setStatus] = useState<Status>(null);
  const [pubKey, setPubKey] = useState<string>('');
  const [roster, setRoster] = useState<string[]>(null);
  const [result, setResult] = useState<Results[]>(null);
  const [chunksPerBallot, setChunksPerBallot] = useState<number>(0);
  const [ballotSize, setBallotSize] = useState<number>(0);
  const [configObj, setConfigObj] = useState(null);
  const [voters, setVoters] = useState<string[]>(null);
  const [isResultSet, setIsResultSet] = useState<boolean>(false);

  useEffect(() => {
    if (formData !== null) {
      const title = JSON.parse(formData.Configuration.MainTitle);
      formData.Configuration.TitleEn = title.en;
      formData.Configuration.TitleFr = title.fr;
      formData.Configuration.TitleDe = title.de;  
      console.log('ok')
      setId(formData.FormID);
      setStatus(formData.Status);
      setPubKey(formData.Pubkey);
      setRoster(formData.Roster);
      setResult(formData.Result);
      setChunksPerBallot(formData.ChunksPerBallot);
      setBallotSize(formData.BallotSize);
      setConfigObj(formData.Configuration);
      setVoters(formData.Voters);

      if (formData.Result.length > 0) {
        setIsResultSet(true);
      }
    }
  }, [formData]);

  return {
    id,
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
  };
};

const useFillLightFormInfo = (formData: LightFormInfo) => {
  const [id, setId] = useState<ID>('');
  const [title, setTitle] = useState<string>('');
  const [status, setStatus] = useState<Status>(null);
  const [pubKey, setPubKey] = useState<string>('');

  useEffect(() => {
    if (formData !== null) {
      setId(formData.FormID);
      setTitle(formData.Title);
      setStatus(formData.Status);
      setPubKey(formData.Pubkey);
    }
  }, [formData]);

  return {
    id,
    title,
    status,
    setStatus,
    pubKey,
  };
};

export { useFillFormInfo, useFillLightFormInfo };
