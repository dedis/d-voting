import React, { FC, useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

import useElection from 'components/utils/useElection';
import useGetResults from 'components/utils/useGetResults';
import { OngoingAction, Status } from 'types/election';
import Modal from 'components/modal/Modal';
import StatusTimeline from './components/StatusTimeline';
import Loading from 'pages/Loading';
import * as endpoints from '../../components/utils/Endpoints';
import useFetchCall from '../../components/utils/useFetchCall';
import Action from './components/Action';
import { NodeStatus } from 'types/node';
import DKGTable from './components/DKGTable';

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

  const request = {
    method: 'GET',
  };
  const [nodeProxyObject, nodeProxyLoading, nodeProxyError] = useFetchCall(
    endpoints.getProxiesAddresses(electionId),
    request
  );
  const [nodeProxyAddresses, setNodeProxyAddresses] = useState<Map<string, string>>(null);
  // The status of each node
  const [DKGStatuses, setDKGStatuses] = useState<Map<string, NodeStatus>>(null);

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

    if (storedOngoingAction) {
      setOngoingAction(storedOngoingAction);
    }
  }, []);

  // Store the ongoingAction in the local storage
  useEffect(() => {
    window.localStorage.setItem(ongoingItem, ongoingAction.toString());
  }, [ongoingAction]);

  // Set the mapping of the node and proxy addresses
  useEffect(() => {
    if (nodeProxyError !== null) {
      setTextModalError(nodeProxyError.message);
      setShowModalError(true);
    }

    if (nodeProxyObject !== null) {
      const newNodeProxyAddresses = new Map();

      nodeProxyObject.Proxies.forEach((value) => {
        Object.entries(value).forEach((v) => {
          newNodeProxyAddresses.set(v[0], v[1]);
          console.log(v[0]);
        });
      });

      setNodeProxyAddresses(newNodeProxyAddresses);
    }
  }, [nodeProxyObject, nodeProxyError]);

  // Fetch the status of the nodes
  useEffect(() => {
    const fetchData = async (node: string) => {
      try {
        console.log(nodeProxyAddresses.get(node));
        const response = await fetch(
          endpoints.getDKGActors(nodeProxyAddresses.get(node), electionId),
          request
        );
        if (!response.ok) {
          if (response.status === 404) {
            return { id: node, status: NodeStatus.NotInitialized };
          } else {
            const js = await response.json();
            throw new Error(JSON.stringify(js));
          }
        } else {
          let dataReceived = await response.json();
          return { id: node, status: dataReceived.Status };
        }
      } catch (e) {
        setTextModalError(e.message);
        setShowModalError(true);
      }
    };

    if (nodeProxyAddresses !== null) {
      const promises: Promise<{
        id: string;
        status: any;
      }>[] = Array.from(nodeProxyAddresses.keys()).map((node) => {
        return fetchData(node);
      });

      Promise.all(promises).then((values) => {
        const newDKGStatuses = new Map();
        values.forEach((v) => newDKGStatuses.set(v.id, v.status));
        setDKGStatuses(newDKGStatuses);
      });
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [nodeProxyAddresses]);

  // Update the status of the election if necessary
  useEffect(() => {
    if (DKGStatuses !== null) {
      if (status === Status.Initial) {
        const statuses = Array.from(DKGStatuses.values());

        if (!statuses.includes(NodeStatus.NotInitialized)) {
          if (statuses.includes(NodeStatus.Setup)) {
            setStatus(Status.Setup);
          } else {
            setStatus(Status.Initialized);
          }
        }
        // Status Failed is handled by useChangeAction
      }
    }
  }, [DKGStatuses, status]);

  return (
    <div className="w-[60rem] font-sans px-4 py-4">
      {!loading ? (
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
              <>
                {!nodeProxyLoading ? (
                  <DKGTable
                    nodeProxyAddresses={nodeProxyAddresses}
                    setNodeProxyAddresses={setNodeProxyAddresses}
                    DKGStatuses={DKGStatuses}
                    electionID={electionId}
                    setTextModalError={setTextModalError}
                    setShowModalError={setShowModalError}
                  />
                ) : null}
              </>
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
