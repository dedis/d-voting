import React, { FC, useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

import useElection from 'components/utils/useElection';
import useGetResults from 'components/utils/useGetResults';
import { NodeStatus, OngoingAction, Status } from 'types/election';
import Modal from 'components/modal/Modal';
import StatusTimeline from './components/StatusTimeline';
import Loading from 'pages/Loading';
import * as endpoints from '../../components/utils/Endpoints';
import useFetchCall from '../../components/utils/useFetchCall';
import Action from './components/Action';

const ElectionShow: FC = () => {
  const { t } = useTranslation();
  const { electionId } = useParams();

  const { loading, electionID, status, setStatus, roster, setResult, configObj, setIsResultSet } =
    useElection(electionId);

  const [, setError] = useState(null);
  const [isResultAvailable, setIsResultAvailable] = useState(false);
  const { getResults } = useGetResults();

  const [getError, setGetError] = useState(null);
  const [textModalError, setTextModalError] = useState(null);
  const [showModalError, setShowModalError] = useState(false);

  const [ongoingAction, setOngoingAction] = useState(OngoingAction.None);
  const request = {
    method: 'GET',
  };
  const [dkgStatus, dkgLoading, dkgError] = useFetchCall(
    endpoints.editDKGActors(electionId),
    request
  );
  const ongoingItem = 'ongoingAction' + electionID;

  // Fetch result when available after a status change
  useEffect(() => {
    if (status === Status.ResultAvailable && isResultAvailable) {
      getResults(electionID, setError, setResult, setIsResultSet);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isResultAvailable, status]);

  useEffect(() => {
    if (getError !== null) {
      console.log(getError);
      setTextModalError(getError);
      setShowModalError(true);
      setGetError(null);
    }
  }, [getError, setTextModalError]);

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

  // Set the ongoingAction in the storage
  useEffect(() => {
    window.localStorage.setItem(ongoingItem, ongoingAction.toString());
  }, [ongoingAction]);

  // Fetch the status of the nodes
  useEffect(() => {
    if (dkgError !== null) {
      // Ignore error 404 when node is not initialized
      if (!dkgError.message.includes('election does not exist')) {
        setTextModalError(dkgError.message + ' show');
        setShowModalError(true);
      }
    } else {
      if (status == Status.Initial) {
        if ((dkgStatus.Status as NodeStatus) === NodeStatus.Initialized) {
          setStatus(Status.Initialized);
        }
        if ((dkgStatus.Status as NodeStatus) === NodeStatus.Setup) {
          setStatus(Status.Setup);
        }
        // Status Failed is handled by useChangeAction
      }
    }
  }, [dkgStatus, status, dkgError, ongoingAction]);

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
                setStatus={setStatus}
                setResultAvailable={setIsResultAvailable}
                setGetError={setGetError}
                setTextModalError={setTextModalError}
                setShowModalError={setShowModalError}
                ongoingAction={ongoingAction}
                setOngoingAction={setOngoingAction}
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
