import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';

import ConfirmModal from '../modal/ConfirmModal';
import { ROUTE_ELECTION_SHOW } from '../../Routes';
import usePostCall from './usePostCall';
import {
  ENDPOINT_EVOTING_CANCEL,
  ENDPOINT_EVOTING_CLOSE,
  ENDPOINT_EVOTING_DECRYPT,
  ENDPOINT_EVOTING_SHUFFLE,
} from './Endpoints';
import { CANCELED, CLOSED, OPEN, RESULT_AVAILABLE, SHUFFLED_BALLOT } from './StatusNumber';
import { COLLECTIVE_AUTHORITY_MEMBERS } from './CollectiveAuthorityMembers';

const useChangeAction = (
  status: number,
  electionID: number,
  setStatus: (status: number) => void,
  setResultAvailable: ((available: boolean) => void | null) | undefined,
  setTextModalError: (value: ((prevState: null) => '') | string) => void,
  setShowModalError: (willShow: boolean) => void
) => {
  const { t } = useTranslation();
  const userID = sessionStorage.getItem('id');
  const token = sessionStorage.getItem('token');
  const [isClosing, setIsClosing] = useState(false);
  const [isCanceling, setIsCanceling] = useState(false);
  const [isShuffling, setIsShuffling] = useState(false);
  const [isDecrypting, setIsDecrypting] = useState(false);
  const [showModalClose, setShowModalClose] = useState(false);
  const [showModalCancel, setShowModalCancel] = useState(false);
  const [userConfirmedClosing, setUserConfirmedClosing] = useState(false);
  const [userConfirmedCanceling, setUserConfirmedCanceling] = useState(false);
  const modalClose = (
    <ConfirmModal
      showModal={showModalClose}
      setShowModal={setShowModalClose}
      textModal={t('confirmCloseElection')}
      setUserConfirmedAction={setUserConfirmedClosing}
    />
  );
  const modalCancel = (
    <ConfirmModal
      showModal={showModalCancel}
      setShowModal={setShowModalCancel}
      textModal={t('confirmCancelElection')}
      setUserConfirmedAction={setUserConfirmedCanceling}
    />
  );
  const [postError, setPostError] = useState(t('operationFailure') as string);
  const { postData } = usePostCall(setPostError);
  const simplePostRequest = {
    method: 'POST',
    body: JSON.stringify({
      ElectionID: electionID,
      UserId: userID,
      Token: token,
    }),
  };
  const shuffleRequest = {
    method: 'POST',
    body: JSON.stringify({
      ElectionID: electionID,
      UserId: userID,
      Token: token,
      Members: COLLECTIVE_AUTHORITY_MEMBERS,
    }),
  };

  useEffect(() => {
    if (postError !== null) {
      setTextModalError(postError);
      setPostError(null);
    }
  }, [postError, setTextModalError]);

  useEffect(() => {
    //check if close button was clicked and the user validated the confirmation window
    if (isClosing && userConfirmedClosing) {
      const close = async () => {
        const closeSuccess = await postData(
          ENDPOINT_EVOTING_CLOSE,
          simplePostRequest,
          setIsClosing
        );

        if (closeSuccess) {
          setStatus(CLOSED);
        } else {
          setShowModalError(true);
        }
        setUserConfirmedClosing(false);
      };

      close().catch(console.error);
    }
  }, [
    isClosing,
    postData,
    setShowModalError,
    setStatus,
    simplePostRequest,
    showModalClose,
    userConfirmedClosing,
  ]);

  useEffect(() => {
    if (isCanceling && userConfirmedCanceling) {
      const cancel = async () => {
        const cancelSuccess = await postData(
          ENDPOINT_EVOTING_CANCEL,
          simplePostRequest,
          setIsCanceling
        );
        if (cancelSuccess) {
          setStatus(CANCELED);
        } else {
          setShowModalError(true);
        }
        setUserConfirmedCanceling(false);
        setPostError(null);
      };

      cancel().catch(console.error);
    }
  }, [
    isCanceling,
    postData,
    setShowModalError,
    setStatus,
    simplePostRequest,
    userConfirmedCanceling,
  ]);

  const handleClose = () => {
    setShowModalClose(true);
    setIsClosing(true);
  };

  const handleCancel = () => {
    setShowModalCancel(true);
    setIsCanceling(true);
  };

  const handleShuffle = async () => {
    setIsShuffling(true);
    const shuffleSuccess = await postData(ENDPOINT_EVOTING_SHUFFLE, shuffleRequest, setIsShuffling);
    if (shuffleSuccess && postError === null) {
      setStatus(SHUFFLED_BALLOT);
    } else {
      setShowModalError(true);
      setIsShuffling(false);
    }
    setPostError(null);
  };

  const handleDecrypt = async () => {
    const decryptSucess = await postData(
      ENDPOINT_EVOTING_DECRYPT,
      simplePostRequest,
      setIsDecrypting
    );
    if (decryptSucess && postError === null) {
      if (setResultAvailable !== null) {
        setResultAvailable(true);
      }
      setStatus(RESULT_AVAILABLE);
    } else {
      setShowModalError(true);
      setIsDecrypting(false);
    }
    setPostError(null);
  };

  const getAction = () => {
    switch (status) {
      case OPEN:
        return (
          <span>
            <button id="close-button" className="election-btn" onClick={handleClose}>
              {t('close')}
            </button>
            <button className="election-btn" onClick={handleCancel}>
              {t('cancel')}
            </button>
          </span>
        );
      case CLOSED:
        return (
          <span>
            {isShuffling ? (
              <p className="loading">{t('shuffleOnGoing')}</p>
            ) : (
              <span>
                <button className="election-btn" onClick={handleShuffle}>
                  {t('shuffle')}
                </button>
              </span>
            )}
          </span>
        );
      case SHUFFLED_BALLOT:
        return (
          <span>
            {isDecrypting ? (
              <p className="loading">{t('decryptOnGoing')}</p>
            ) : (
              <span>
                <button className="election-btn" onClick={handleDecrypt}>
                  {t('decrypt')}
                </button>
              </span>
            )}
          </span>
        );
      case RESULT_AVAILABLE:
        return (
          <span>
            <Link
              className="election-link"
              to={{ pathname: `${ROUTE_ELECTION_SHOW}/${electionID}` }}>
              <button className="election-btn">{t('seeResult')}</button>
            </Link>
          </span>
        );
      case CANCELED:
        return <span> ---</span>;
      default:
        return <span> --- </span>;
    }
  };
  return { getAction, modalClose, modalCancel };
};

export default useChangeAction;
