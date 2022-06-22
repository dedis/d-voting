import React, { FC, useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

import useElection from 'components/utils/useElection';
import { OngoingAction, Status } from 'types/election';
import Modal from 'components/modal/Modal';
import StatusTimeline from './components/StatusTimeline';
import Loading from 'pages/Loading';
import * as endpoints from '../../components/utils/Endpoints';
import Action from './components/Action';
import { NodeProxyAddress, NodeStatus } from 'types/node';
import useGetResults from './components/utils/useGetResults';
import DKGStatus from 'components/utils/DKGStatus';

const ElectionShow: FC = () => {
  const { t } = useTranslation();
  const { electionId } = useParams();
  const {
    loading,
    electionID,
    status,
    setStatus,
    roster,
    setResult,
    configObj,
    setIsResultSet,
    error,
  } = useElection(electionId);

  const [, setError] = useState(null);
  const [isResultAvailable, setIsResultAvailable] = useState(false);
  const { getResults } = useGetResults();

  const [textModalError, setTextModalError] = useState(null);
  const [showModalError, setShowModalError] = useState(false);

  const [ongoingAction, setOngoingAction] = useState(OngoingAction.None);

  const [nodeProxyAddresses, setNodeProxyAddresses] = useState<Map<string, string>>(null);
  const [nodeToSetup, setNodeToSetup] = useState<[string, string]>(null);
  // The status of each node
  const [DKGStatuses, setDKGStatuses] = useState<Map<string, NodeStatus>>(null);
  const [DKGLoading, setDKGLoading] = useState(true);
  const request = {
    method: 'GET',
    headers: {
      'Access-Control-Allow-Origin': '*',
    },
  };

  const ongoingItem = 'ongoingAction' + electionID;
  const nodeToSetupItem = 'nodeToSetup' + electionID;

  // Fetch result when available after a status change
  useEffect(() => {
    if (status === Status.ResultAvailable && isResultAvailable) {
      getResults(electionID, setError, setResult, setIsResultSet);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isResultAvailable, status]);

  // Clean up the storage when it's not needed anymore
  useEffect(() => {
    if (status === Status.ResultAvailable) {
      window.localStorage.removeItem(ongoingItem);
    }

    if (status === Status.Setup) {
      window.localStorage.removeItem(nodeToSetupItem);
    }
  }, [status]);

  // Get the ongoingAction and the nodeToSetup from the storage
  useEffect(() => {
    const storedOngoingAction = JSON.parse(window.localStorage.getItem(ongoingItem));

    if (storedOngoingAction !== null) {
      setOngoingAction(storedOngoingAction);
    }

    const storedNodeToSetup = JSON.parse(window.localStorage.getItem(nodeToSetupItem));

    if (storedNodeToSetup !== null) {
      setNodeToSetup([storedNodeToSetup[0], storedNodeToSetup[1]]);
    }
  }, []);

  // Store the ongoingAction and the nodeToSetup in the local storage
  useEffect(() => {
    if (status !== Status.ResultAvailable) {
      window.localStorage.setItem(ongoingItem, ongoingAction.toString());
    }
  }, [ongoingAction]);

  useEffect(() => {
    if (nodeToSetup !== null) {
      window.localStorage.setItem(nodeToSetupItem, JSON.stringify(nodeToSetup));
    }
  }, [nodeToSetup]);

  useEffect(() => {
    if (nodeProxyAddresses !== null) {
      if (nodeToSetup === null) {
        const node = roster[0];
        setNodeToSetup([node, nodeProxyAddresses.get(node)]);
      }
    }
  }, [nodeProxyAddresses, nodeToSetup]);

  // Set the mapping of the node and proxy addresses
  useEffect(() => {
    if (roster !== null) {
      const fetchNodeProxy = async (node: string) => {
        try {
          const response = await fetch(endpoints.getProxyAddress(node), request);

          if (!response.ok) {
            const js = await response.json();
            throw new Error(JSON.stringify(js));
          } else {
            let dataReceived = await response.json();
            return dataReceived as NodeProxyAddress;
          }
        } catch (e) {
          setTextModalError(t('errorRetrievingProxy') + e.message);
          setShowModalError(true);
          return Promise.reject();
        }
      };

      const promise = roster.map((node) => {
        return fetchNodeProxy(node);
      });

      Promise.all(promise).then(
        (nodeProxies) => {
          const newAddresses: Map<string, string> = new Map();

          nodeProxies.forEach((nodeProxy) => {
            newAddresses.set(nodeProxy.NodeAddr, nodeProxy.Proxy);
          });

          setNodeProxyAddresses(newAddresses);
        },
        () => {
          setDKGLoading(false);
        }
      );
    }
  }, [roster]);

  // Fetch the status of the nodes
  useEffect(() => {
    if (nodeProxyAddresses !== null) {
      const fetchDKGStatus = async (node: string, proxy: string) => {
        try {
          const response = await fetch(endpoints.getDKGActors(proxy, electionId), request);

          if (response.status === 404) {
            return { NodeAddr: node, Status: NodeStatus.NotInitialized };
          }

          if (!response.ok) {
            const js = await response.json();
            throw new Error(JSON.stringify(js));
          }

          let dataReceived = await response.json();
          return { NodeAddr: node, Status: dataReceived.Status };
        } catch (e) {
          setTextModalError(t('errorRetrievingNodes') + e.message);
          setShowModalError(true);

          return Promise.reject();
        }
      };

      const promises = Array.from(nodeProxyAddresses).map(([node, proxy]) => {
        return fetchDKGStatus(node, proxy);
      });

      Promise.all(promises)
        .then(
          (values) => {
            const newDKGStatuses = new Map();
            values.forEach((v) => newDKGStatuses.set(v.NodeAddr, v.Status));
            setDKGStatuses(newDKGStatuses);
          },
          () => {
            setDKGLoading(false);
          }
        )
        .finally(() => setDKGLoading(false));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [nodeProxyAddresses]);

  // Update the status of the election if necessary
  useEffect(() => {
    if (DKGStatuses !== null) {
      if (status === Status.Initial) {
        const statuses = Array.from(DKGStatuses.values());

        if (statuses.includes(NodeStatus.NotInitialized)) return;

        if (statuses.includes(NodeStatus.Setup)) {
          setStatus(Status.Setup);
        } else {
          setStatus(Status.Initialized);
        }
        // Status Failed is handled by useChangeAction
      }
    }
  }, [DKGStatuses, status]);

  useEffect(() => {
    if (error !== null) {
      setTextModalError(t('errorRetrievingElection') + error.message);
      setShowModalError(true);
      setError(null);
    }
  }, [error]);

  return (
    <div className="w-[60rem] font-sans px-4 py-4">
      <Modal
        showModal={showModalError}
        setShowModal={setShowModalError}
        textModal={textModalError === null ? '' : textModalError}
        buttonRightText={t('close')}
      />
      {!loading && !DKGLoading ? (
        <>
          <div className="pt-8 text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
            {configObj.MainTitle}
          </div>

          <div className="break-all">Election ID : {electionId}</div>
          <div className="py-6 pl-2">
            <div className="font-bold uppercase text-lg text-gray-700">{t('status')}</div>

            <div className="px-2 pt-6 flex justify-center">
              <StatusTimeline status={status} ongoingAction={ongoingAction} />
            </div>
          </div>
          <div className="py-4 pl-2 pb-8">
            <div className="font-bold uppercase text-lg text-gray-700 pb-2">{t('action')}</div>
            <div className="px-2">
              <Action
                status={status}
                electionID={electionID}
                roster={roster}
                nodeProxyAddresses={nodeProxyAddresses}
                setStatus={setStatus}
                setResultAvailable={setIsResultAvailable}
                setTextModalError={setTextModalError}
                setShowModalError={setShowModalError}
                ongoingAction={ongoingAction}
                setOngoingAction={setOngoingAction}
                nodeToSetup={nodeToSetup}
                setNodeToSetup={setNodeToSetup}
                DKGStatuses={DKGStatuses}
                setDKGStatuses={setDKGStatuses}
              />
            </div>
          </div>
          <div className="py-4 pl-2 pb-8">
            <div className="font-bold uppercase text-lg text-gray-700 pb-2">{t('DKGStatuses')}</div>
            <div className="px-2">
              {roster.map((node, index) => (
                <div className="flex flex-col pb-6" key={node}>
                  {t('node')} {index} ({node})
                  {nodeProxyAddresses !== null && DKGStatuses !== null ? (
                    <DKGStatus status={DKGStatuses.get(node)} />
                  ) : (
                    <DKGStatus status={NodeStatus.NotInitialized} />
                  )}
                </div>
              ))}
            </div>
          </div>
        </>
      ) : (
        <Loading />
      )}
    </div>
  );
};

ElectionShow.propTypes = {
  location: PropTypes.any,
};

export default ElectionShow;
