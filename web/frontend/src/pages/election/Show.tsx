import React, { FC, useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

import useElection from 'components/utils/useElection';
import { OngoingAction, Status } from 'types/election';
import Modal from 'components/modal/Modal';
import StatusTimeline from './components/StatusTimeline';
import Loading from 'pages/Loading';
import Action from './components/Action';
import { NodeStatus } from 'types/node';
import useGetResults from './components/utils/useGetResults';
import UserIDTable from './components/UserIDTable';
import DKGStatusTable from './components/DKGStatusTable';
import LoadingButton from './components/LoadingButton';

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
    voters,
    error,
  } = useElection(electionId);

  const [, setError] = useState(null);
  const [isResultAvailable, setIsResultAvailable] = useState(false);
  const { getResults } = useGetResults();

  const [textModalError, setTextModalError] = useState(null);
  const [showModalError, setShowModalError] = useState(false);

  const [ongoingAction, setOngoingAction] = useState(OngoingAction.None);

  const [nodeProxyAddresses, setNodeProxyAddresses] = useState<Map<string, string>>(new Map());
  const [nodeToSetup, setNodeToSetup] = useState<[string, string]>(null);
  // The status of each node
  const [DKGStatuses, setDKGStatuses] = useState<Map<string, NodeStatus>>(new Map());

  const [nodeLoading, setNodeLoading] = useState<Map<string, boolean>>(null);
  const [DKGLoading, setDKGLoading] = useState(true);

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
    // eslint-disable-next-line react-hooks/exhaustive-deps
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
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Store the ongoingAction and the nodeToSetup in the local storage
  useEffect(() => {
    if (status !== Status.ResultAvailable) {
      window.localStorage.setItem(ongoingItem, ongoingAction.toString());
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ongoingAction]);

  useEffect(() => {
    if (nodeToSetup !== null) {
      window.localStorage.setItem(nodeToSetupItem, JSON.stringify(nodeToSetup));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [nodeToSetup]);

  useEffect(() => {
    // Set default node to initialize
    if (status >= Status.Initialized) {
      setNodeToSetup(Array.from(nodeProxyAddresses).find(([_node, proxy]) => proxy !== ''));
    }
  }, [nodeProxyAddresses, status]);

  useEffect(() => {
    if (roster !== null) {
      const newNodeLoading = new Map();
      roster.forEach((node) => {
        newNodeLoading.set(node, true);
      });

      setNodeLoading(newNodeLoading);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [roster]);

  useEffect(() => {
    if (nodeLoading !== null) {
      if (!Array.from(nodeLoading.values()).includes(true)) {
        setDKGLoading(false);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [nodeLoading]);

  // Update the status of the election if necessary
  useEffect(() => {
    if (status === Status.Initial) {
      if (DKGStatuses !== null && !DKGLoading) {
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

    if (status >= Status.Open) {
      setDKGLoading(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [DKGStatuses, status, DKGLoading]);

  useEffect(() => {
    if (error !== null) {
      setTextModalError(t('errorRetrievingElection') + error.message);
      setShowModalError(true);
      setError(null);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [error]);

  return (
    <div className="w-[60rem] font-sans px-4 py-4">
      <Modal
        showModal={showModalError}
        setShowModal={setShowModalError}
        textModal={textModalError === null ? '' : textModalError}
        buttonRightText={t('close')}
      />
      {!loading ? (
        <>
          <div className="pt-8 text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
            {configObj.MainTitle}
          </div>

          <div className="pt-2 break-all">Election ID : {electionId}</div>
          {status >= Status.Open && status <= Status.Canceled && (
            <div className="break-all">{t('numVotes', { num: voters.length })}</div>
          )}
          <div className="py-6 pl-2">
            <div className="font-bold uppercase text-lg text-gray-700">{t('status')}</div>
            {DKGLoading && (
              <div className="px-2 pt-6">
                <LoadingButton>{t('statusLoading')}</LoadingButton>
              </div>
            )}
            {!DKGLoading && (
              <div className="px-2 pt-6 flex justify-center">
                <StatusTimeline status={status} ongoingAction={ongoingAction} />
              </div>
            )}
          </div>
          <div className="py-4 pl-2 pb-8">
            <div className="font-bold uppercase text-lg text-gray-700 pb-2">{t('action')}</div>
            <div className="px-2">
              {DKGLoading && <LoadingButton>{t('actionLoading')}</LoadingButton>}{' '}
              {!DKGLoading && (
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
              )}
            </div>
          </div>
          {voters.length > 0 && (
            <div className="py-4 pl-2 pb-8">
              <div className="font-bold uppercase text-lg text-gray-700 pb-2">{t('userID')}</div>
              <div className="px-2">
                <UserIDTable userIDs={voters} />
              </div>
            </div>
          )}
          <div className="py-4 pl-2 pb-8">
            <div className="font-bold uppercase text-lg text-gray-700 pb-2">{t('DKGStatuses')}</div>
            <div className="px-2">
              <DKGStatusTable
                roster={roster}
                electionId={electionId}
                loading={nodeLoading}
                setLoading={setNodeLoading}
                nodeProxyAddresses={nodeProxyAddresses}
                setNodeProxyAddresses={setNodeProxyAddresses}
                DKGStatuses={DKGStatuses}
                setDKGStatuses={setDKGStatuses}
                setTextModalError={setTextModalError}
                setShowModalError={setShowModalError}
              />
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
