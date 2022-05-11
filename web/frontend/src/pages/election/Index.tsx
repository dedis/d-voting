import React, { FC, useContext, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import ElectionTable from './components/ElectionTable';
import useFetchCall from 'components/utils/useFetchCall';
import * as endpoints from 'components/utils/Endpoints';
import Loading from 'pages/Loading';
import { LightElectionInfo, NodeStatus, Status } from 'types/election';
import ElectionTableFilter from './components/ElectionTableFilter';
import { FlashContext, FlashLevel } from 'index';
import { ID } from 'types/configuration';

const ElectionIndex: FC = () => {
  const { t } = useTranslation();
  const fctx = useContext(FlashContext);
  const [statusToKeep, setStatusToKeep] = useState<Status>(null);
  const [elections, setElections] = useState<LightElectionInfo[]>(null);
  const [DKGStatuses, setDKGStatuses] = useState<Map<ID, Status>>(new Map());
  const [electionStatuses, setElectionsStatuses] = useState<Map<ID, Status>>(new Map());

  const request = {
    method: 'GET',
    headers: {
      'Access-Control-Allow-Origin': '*',
    },
  };

  const [data, loading, error] = useFetchCall(endpoints.elections, request);

  // Apply the filter statusToKeep
  useEffect(() => {
    if (data !== null) {
      if (statusToKeep !== null) {
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
      } else {
        setElections(data.Elections);
      }
    }
  }, [data, statusToKeep]);

  // Get the node status for each election
  useEffect(() => {
    if (elections !== null) {
      const req = {
        method: 'GET',
      };
      const fetchData = async (election: LightElectionInfo) => {
        try {
          const response = await fetch(endpoints.getDKGActors(election.ElectionID), req);
          if (!response.ok) {
            // The node is not initialized
            if (response.status === 404) {
              return { ID: election.ElectionID, Status: Status.Initial };
            } else {
              const js = await response.json();
              fctx.addMessage(JSON.stringify(js), FlashLevel.Error);
            }
          } else {
            let dkgStatus = await response.json();

            if ((dkgStatus.Status as NodeStatus) === NodeStatus.Initialized) {
              return { ID: election.ElectionID, Status: Status.Initialized };
            }
            if ((dkgStatus.Status as NodeStatus) === NodeStatus.Setup) {
              return { ID: election.ElectionID, Status: Status.Setup };
            }
            // NodeStatus Failed is handled in useChangeAction
          }
        } catch (e) {
          fctx.addMessage(e.message, FlashLevel.Error);
        }
      };

      const newDKGStatuses = new Map(DKGStatuses);

      const promises: Promise<{
        ID: string;
        Status: Status;
      }>[] = elections.map((election) => {
        return fetchData(election);
      });

      Promise.all(promises).then((values) => {
        values.forEach((v) => newDKGStatuses.set(v.ID, v.Status));
        setDKGStatuses(newDKGStatuses);
      });
    }
  }, [elections]);

  // Set the node status for each election
  useEffect(() => {
    if (data !== null) {
      const newElectionStatuses = new Map(electionStatuses);

      elections.forEach((election) => {
        newElectionStatuses.set(election.ElectionID, election.Status);

        if (election.Status === Status.Initial) {
          if (DKGStatuses.get(election.ElectionID)) {
            newElectionStatuses.set(election.ElectionID, DKGStatuses.get(election.ElectionID));
          }
        }
      });
      setElectionsStatuses(newElectionStatuses);
    }
  }, [elections, DKGStatuses]);

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
          <ElectionTable elections={elections} electionStatuses={electionStatuses} />
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
