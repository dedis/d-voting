import React, { FC, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import ElectionTable from './components/ElectionTable';
import useFetchCall from 'components/utils/useFetchCall';
import * as endpoints from 'components/utils/Endpoints';
import Loading from 'pages/Loading';
import { LightElectionInfo, Status } from 'types/election';
import ElectionTableFilter from './components/ElectionTableFilter';
import { ID } from 'types/configuration';

const ElectionIndex: FC = () => {
  const { t } = useTranslation();

  const [statusToKeep, setStatusToKeep] = useState<Status>(null);
  const [elections, setElections] = useState<LightElectionInfo[]>(null);
  const [electionStatuses, setElectionsStatuses] = useState<Map<ID, Status>>(new Map());

  const request = {
    method: 'GET',
    headers: {
      'Access-Control-Allow-Origin': '*',
    },
  };

  const [data, loading, error] = useFetchCall(endpoints.elections, request);

  useEffect(() => {
    if (data !== null) {
      const newStatuses = new Map(electionStatuses);
      (data.Elections as LightElectionInfo[]).forEach((election) => {
        newStatuses.set(election.ElectionID, election.Status);
      });

      setElectionsStatuses(newStatuses);
    }
  }, [data]);

  // Apply the filter statusToKeep
  useEffect(() => {
    if (data === null) return;

    if (statusToKeep === null) {
      setElections(data.Elections);
      return;
    }

    const filteredElectionsID = [];
    electionStatuses.forEach((status, id) => {
      if (status === statusToKeep) {
        filteredElectionsID.push(id);
      }
    });

    const filteredElections = (data.Elections as LightElectionInfo[]).filter((election) =>
      filteredElectionsID.includes(election.ElectionID)
    );

    setElections(filteredElections);
  }, [data, statusToKeep]);

  return (
    <div className="w-[60rem] font-sans px-4 py-4">
      {!loading ? (
        <div className="py-8">
          <h2 className="pb-2 text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
            {t('elections')}
          </h2>
          <div>{t('listElection')}</div>
          <div>{t('clickElection')}</div>
          <ElectionTableFilter setStatusToKeep={setStatusToKeep} />
          <ElectionTable
            elections={elections}
            electionStatuses={electionStatuses}
            setElectionsStatuses={setElectionsStatuses}
          />
        </div>
      ) : error === null ? (
        <Loading />
      ) : (
        <div>
          {t('errorRetrievingElection')} - {error.toString()}
        </div>
      )}
    </div>
  );
};

export default ElectionIndex;
