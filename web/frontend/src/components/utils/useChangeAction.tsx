import React, { useContext, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import ConfirmModal from '../modal/ConfirmModal';
import usePostCall from './usePostCall';
import * as endpoints from './Endpoints';
import { ID } from 'types/configuration';
import { STATUS } from 'types/electionInfo';
import { AuthContext } from 'index';

const useChangeAction = (
  status: STATUS,
  electionID: ID,
  setStatus: (status: STATUS) => void,
  setResultAvailable: ((available: boolean) => void | null) | undefined,
  setTextModalError: (value: ((prevState: null) => '') | string) => void,
  setShowModalError: (willShow: boolean) => void
) => {
  const { t } = useTranslation();
  const { role } = useContext(AuthContext);
  const [isOpening, setIsOpening] = useState(false);
  const [isClosing, setIsClosing] = useState(false);
  const [isCanceling, setIsCanceling] = useState(false);
  const [isShuffling, setIsShuffling] = useState(false);
  const [isDecrypting, setIsDecrypting] = useState(false);
  const [showModalOpen, setShowModalOpen] = useState(false);
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
  const sendFetchRequest = usePostCall(setPostError);

  const electionUpdate = async (action: string, endpoint: string) => {
    const req = {
      method: 'PUT',
      body: JSON.stringify({
        Action: action,
      }),
      headers: {
        'Content-Type': 'application/json',
      },
    };
    return sendFetchRequest(endpoint, req, setIsClosing);
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
        const closeSuccess = await electionUpdate(
          'close',
          endpoints.editElection(electionID.toString())
        );
        if (closeSuccess) {
          setStatus(STATUS.CLOSED);
        } else {
          setShowModalError(true);
        }
        setUserConfirmedClosing(false);
      };

      close().catch(console.error);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    isClosing,
    sendFetchRequest,
    setShowModalError,
    setStatus,
    showModalClose,
    userConfirmedClosing,
  ]);

  useEffect(() => {
    if (isCanceling && userConfirmedCanceling) {
      const cancel = async () => {
        const cancelSuccess = await electionUpdate(
          'cancel',
          endpoints.editElection(electionID.toString())
        );
        if (cancelSuccess) {
          setStatus(STATUS.CANCELED);
        } else {
          setShowModalError(true);
        }
        setUserConfirmedCanceling(false);
        setPostError(null);
      };

      cancel().catch(console.error);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isCanceling, sendFetchRequest, setShowModalError, setStatus, userConfirmedCanceling]);

  const handleOpen = async () => {
    const openSuccess = await electionUpdate('open', endpoints.editElection(electionID.toString()));
    if (openSuccess && postError === null) {
      setStatus(STATUS.OPEN);
    } else {
      setShowModalError(true);
      setIsOpening(false);
    }
    setPostError(null);
  };

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
    const shuffleSuccess = await electionUpdate(
      'shuffle',
      endpoints.editShuffle(electionID.toString())
    );
    if (shuffleSuccess && postError === null) {
      setStatus(STATUS.SHUFFLED_BALLOTS);
    } else {
      setShowModalError(true);
      setIsShuffling(false);
    }
    setPostError(null);
  };

  const handleDecrypt = async () => {
    const decryptSuccess = await electionUpdate(
      'beginDecryption',
      endpoints.editDKGActors(electionID.toString())
    );
    if (decryptSuccess && postError === null) {
      // TODO : setResultAvailable is undefined when the decryption is clicked
      if (setResultAvailable !== null && setResultAvailable !== undefined) {
        setResultAvailable(true);
      }
      setStatus(STATUS.RESULT_AVAILABLE);
    } else {
      setShowModalError(true);
      setIsDecrypting(false);
    }
    setPostError(null);
  };

  const getAction = () => {
    if (role === 'admin' || role === 'operator') {
      switch (status) {
        case STATUS.INITIAL:
          return (
            <span>
              <button id="close-button" className="election-btn" onClick={handleOpen}>
                {t('open')}
              </button>
              <button className="election-btn" onClick={handleCancel}>
                {t('cancel')}
              </button>
            </span>
          );
        case STATUS.OPEN:
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
        case STATUS.CLOSED:
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
        case STATUS.SHUFFLED_BALLOTS:
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
        case STATUS.RESULT_AVAILABLE:
          return (
            <span>
              <Link className="election-link" to={`/elections/${electionID}`}>
                <button className="election-btn">{t('seeResult')}</button>
              </Link>
            </span>
          );
        case STATUS.CANCELED:
          return <span> --- </span>;
        default:
          return <span> --- </span>;
      }
    } else if (role === 'user') {
      switch (status) {
        case STATUS.RESULT_AVAILABLE:
          return (
            <span>
              <Link className="election-link" to={`/elections/${electionID}`}>
                <button className="election-btn">{t('seeResult')}</button>
              </Link>
            </span>
          );
        default:
          return <span> --- </span>;
      }
    } else {
      switch (status) {
        // TODO: when the action Vote is added, an unauthenticated user won't
        // see the vote action
        case STATUS.RESULT_AVAILABLE:
          return (
            <span>
              <Link className="election-link" to={`/elections/${electionID}`}>
                <button className="election-btn">{t('seeResult')}</button>
              </Link>
            </span>
          );
        default:
          return <span> --- </span>;
      }
    }
  };
  return { getAction, modalClose, modalCancel };
};

export default useChangeAction;
