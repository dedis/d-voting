import React, { FC, useContext, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import ElectionTable from './components/ElectionTable';
import useFetchCall from 'components/utils/useFetchCall';
import * as endpoints from 'components/utils/Endpoints';
import Loading from 'pages/Loading';
import { LightElectionInfo, Status } from 'types/election';
import ElectionTableFilter from './components/ElectionTableFilter';
import { FlashContext, FlashLevel, ProxyContext } from 'index';

const ElectionIndex: FC = () => {
  const { t } = useTranslation();
  const fctx = useContext(FlashContext);
  const pctx = useContext(ProxyContext);

  const [statusToKeep, setStatusToKeep] = useState<Status>(null);
  const [elections, setElections] = useState<LightElectionInfo[]>(null);
  const [loading, setLoading] = useState(true);

  const request = {
    method: 'GET',
    headers: {
      'Access-Control-Allow-Origin': '*',
    },
  };

  const [data, dataLoading, error] = useFetchCall(endpoints.elections(pctx.getProxy()), request);

  useEffect(() => {
    if (error !== null) {
      fctx.addMessage(t('errorRetrievingElections') + error.message, FlashLevel.Error);
      setLoading(false);
    }
  }, [error]);

  // Apply the filter statusToKeep
  useEffect(() => {
    if (data === null) return;

    if (statusToKeep === null) {
      setElections(data.Elections);
      return;
    }

    const filteredElections = (data.Elections as LightElectionInfo[]).filter(
      (election) => election.Status === statusToKeep
    );

    setElections(filteredElections);
  }, [data, statusToKeep]);

  useEffect(() => {
    if (dataLoading !== null) {
      setLoading(dataLoading);
    }
  }, [dataLoading]);

  return (
    <div className="w-[60rem] font-sans px-4 py-4">
      {!loading ? (
        <div className="py-8">
          <h2 className="pb-2 text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
            {t('elections')}
          </h2>
          <div className="mt-1 flex flex-col sm:mt-0">
            <div className="mt-2 flex items-center text-sm text-gray-500">{t('listElection')}</div>
            <div className="mt-1 flex items-center text-sm text-gray-500">{t('clickElection')}</div>
          </div>

          <ElectionTableFilter setStatusToKeep={setStatusToKeep} />
          <ElectionTable elections={elections} />
        </div>
      ) : (
        <Loading />
      )}
    </div>
  );
};

export default ElectionIndex;
