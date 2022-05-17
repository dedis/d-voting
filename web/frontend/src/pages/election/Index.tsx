import React, { FC, useContext, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import ElectionTable from './components/ElectionTable';
import useFetchCall from 'components/utils/useFetchCall';
import * as endpoints from 'components/utils/Endpoints';
import Loading from 'pages/Loading';
import { LightElectionInfo, Status } from 'types/election';
import ElectionTableFilter from './components/ElectionTableFilter';
import { FlashContext, FlashLevel } from 'index';
import { ID } from 'types/configuration';
import { NodeStatus } from 'types/node';

const ElectionIndex: FC = () => {
  const { t } = useTranslation();
  const fctx = useContext(FlashContext);
  const [statusToKeep, setStatusToKeep] = useState<Status>(null);
  const [elections, setElections] = useState<LightElectionInfo[]>(null);
  const [DKGStatuses, setDKGStatuses] = useState<Map<ID, Map<string, NodeStatus>>>(null);
  const [DKGLoading, setDKGLoading] = useState(true);
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

  // Fetch the NodeStatus for each node of each election
  useEffect(() => {
    if (elections !== null) {
      const fetchDKGStatus = async (id: ID, node: string, proxyAddress: string) => {
        try {
          const response = await fetch(endpoints.getDKGActors(proxyAddress, id), request);
          if (!response.ok) {
            // The node is not initialized
            if (response.status === 404) {
              return { ID: id, Node: node, Status: NodeStatus.NotInitialized };
            } else {
              const js = await response.json();
              fctx.addMessage(JSON.stringify(js), FlashLevel.Error);
            }
          } else {
            let dkgStatus = await response.json();

            return { ID: id, Node: node, Status: dkgStatus.Status as NodeStatus };
          }
        } catch (e) {
          fctx.addMessage(e.message, FlashLevel.Error);
        }
      };

      const fetchProxies = async (election: LightElectionInfo) => {
        try {
          const response = await fetch(endpoints.getProxiesAddresses(election.ElectionID), request);
          if (!response.ok) {
            const js = await response.json();
            throw new Error(JSON.stringify(js));
          } else {
            let dataReceived = await response.json();

            const newNodeProxyAddresses = new Map<string, string>();
            dataReceived.Proxies.forEach((value) => {
              Object.entries(value).forEach(([node, proxy]) => {
                newNodeProxyAddresses.set(node, proxy as string);
              });
            });

            return { ID: election.ElectionID, NodeProxy: newNodeProxyAddresses };
          }
        } catch (e) {
          fctx.addMessage(e.message, FlashLevel.Error);
        }
      };

      const promises = elections
        .filter((election) => election.Status === Status.Initial)
        .map((election) => {
          return fetchProxies(election);
        });

      Promise.all(promises)
        .then((value) => {
          const newNodeProxyAddresses: Map<ID, Map<string, string>> = new Map();
          value.forEach((v) => {
            newNodeProxyAddresses.set(v.ID, v.NodeProxy);
          });
          return newNodeProxyAddresses;
        })
        .then((newNodeProxyAddresses) => {
          return Promise.all(
            Array.from(newNodeProxyAddresses).map(([id, nodeProxy]) => {
              return Promise.all(
                Array.from(nodeProxy).map(([node, proxy]) => {
                  return fetchDKGStatus(id, node, proxy);
                })
              );
            })
          );
        })
        .then((values) => {
          const newDKGStatuses: Map<ID, Map<string, NodeStatus>> = new Map();

          values.forEach((val) => {
            const newDKGStatus: Map<string, NodeStatus> = new Map();

            val.forEach((v) => {
              newDKGStatus.set(v.Node, v.Status);
              newDKGStatuses.set(v.ID, newDKGStatus);
            });
          });

          setDKGStatuses(newDKGStatuses);
        })
        .finally(() => setDKGLoading(false));
    }
  }, [elections]);

  useEffect(() => {
    if (elections !== null && DKGStatuses !== null) {
      const newElectionStatuses = new Map(electionStatuses);

      elections.forEach((election) => {
        newElectionStatuses.set(election.ElectionID, election.Status);

        if (DKGStatuses.get(election.ElectionID)) {
          const dkgStatuses = Array.from(DKGStatuses.get(election.ElectionID).values());

          if (!dkgStatuses.includes(NodeStatus.NotInitialized)) {
            newElectionStatuses.set(election.ElectionID, Status.Initialized);
          }
          if (dkgStatuses.includes(NodeStatus.Setup)) {
            newElectionStatuses.set(election.ElectionID, Status.Setup);
          }
        }
      });
      setElectionsStatuses(newElectionStatuses);
    }
  }, [elections, DKGStatuses]);

  return (
    <div className="w-[60rem] font-sans px-4 py-4">
      {!loading && !DKGLoading ? (
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
