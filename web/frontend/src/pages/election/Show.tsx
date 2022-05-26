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
  const { loading, electionID, status, setStatus, roster, setResult, configObj, setIsResultSet } =
    useElection(electionId);

  const [, setError] = useState(null);
  const [isResultAvailable, setIsResultAvailable] = useState(false);
  const { getResults } = useGetResults();

  const [textModalError, setTextModalError] = useState(null);
  const [showModalError, setShowModalError] = useState(false);

  const [ongoingAction, setOngoingAction] = useState(OngoingAction.None);

  const [nodeProxyAddresses, setNodeProxyAddresses] = useState<Map<string, string>>(null);
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
  }, [status]);

  // Get the ongoingAction from the storage
  useEffect(() => {
    const storedOngoingAction = JSON.parse(window.localStorage.getItem(ongoingItem));

    if (storedOngoingAction !== null) {
      setOngoingAction(storedOngoingAction);
    }
  }, []);

  // Store the ongoingAction in the local storage
  useEffect(() => {
    if (status !== Status.ResultAvailable) {
      window.localStorage.setItem(ongoingItem, ongoingAction.toString());
    }
  }, [ongoingAction]);

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
            console.log(dataReceived);
            return dataReceived as NodeProxyAddress;
          }
        } catch (e) {
          setTextModalError(e.message);
          setShowModalError(true);
        }
      };

      const promise = roster.map((node) => {
        return fetchNodeProxy(node);
      });

      Promise.all(promise).then((nodeProxies) => {
        const newAddresses: Map<string, string> = new Map();

        nodeProxies.forEach((nodeProxy) => {
          newAddresses.set(nodeProxy.NodeAddr, nodeProxy.Proxy);
        });

        setNodeProxyAddresses(newAddresses);
      });
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
          setTextModalError(e.message);
          setShowModalError(true);
        }
      };

      const promises = Array.from(nodeProxyAddresses).map(([node, proxy]) => {
        return fetchDKGStatus(node, proxy);
      });

      Promise.all(promises)
        .then((values) => {
          const newDKGStatuses = new Map();
          values.forEach((v) => newDKGStatuses.set(v.NodeAddr, v.Status));
          setDKGStatuses(newDKGStatuses);
        })
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

  return (
    <div className="w-[60rem] font-sans px-4 py-4">
      {!loading && !DKGLoading ? (
        <>
          <Modal
            showModal={showModalError}
            setShowModal={setShowModalError}
            textModal={textModalError === null ? '' : textModalError}
            buttonRightText={t('close')}
          />
          <h2 className="pt-8 text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
            {configObj.MainTitle}
          </h2>

          <h2>Election ID : {electionId}</h2>
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
                DKGStatuses={DKGStatuses}
                setDKGStatuses={setDKGStatuses}
              />
            </div>
          </div>
          <div className="py-4 pl-2 pb-8">
            <div className="font-bold uppercase text-lg text-gray-700 pb-2">{t('DKGStatuses')}</div>
            <div className="px-2">
              {Array.from(nodeProxyAddresses).map(([node, _proxy], index) => (
                <div className="flex flex-col pb-6" key={node}>
                  {t('node')} {index} ({node})
                  <DKGStatus status={DKGStatuses.get(node)} />
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
