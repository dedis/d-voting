import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import ConfirmModal from '../modal/ConfirmModal';
import usePostCall from './usePostCall';
import * as endpoints from './Endpoints';
import { ID } from 'types/configuration';
import { STATUS } from 'types/election';
import ShuffleButton from './ShuffleButton';
import CloseButton from './CloseButton';
import CancelButton from './CancelButton';
import OpenButton from './OpenButton';
import DecryptButton from './DecryptButton';
import ResultButton from './ResultButton';
import VoteButton from './VoteButton';
import CombineButton from './CombineButton';

const useChangeAction = (
  status: STATUS,
  electionID: ID,
  setStatus: (status: STATUS) => void,
  setResultAvailable: ((available: boolean) => void | null) | undefined,
  setTextModalError: (value: ((prevState: null) => '') | string) => void,
  setShowModalError: (willShow: boolean) => void
) => {
  const { t } = useTranslation();
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
          setStatus(STATUS.Closed);
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
          setStatus(STATUS.Canceled);
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
      setStatus(STATUS.Open);
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
      setStatus(STATUS.ShuffledBallots);
    } else {
      setShowModalError(true);
      setIsShuffling(false);
    }
    setPostError(null);
  };

  const handleDecrypt = async () => {
    const decryptSuccess = await electionUpdate(
      'computePubshares',
      endpoints.editDKGActors(electionID.toString())
    );
    if (decryptSuccess && postError === null) {
      // TODO : setResultAvailable is undefined when the decryption is clicked
      // if (setResultAvailable !== null && setResultAvailable !== undefined) {
      //   setResultAvailable(true);
      // }
      setStatus(STATUS.DecryptedBallots);
    } else {
      setShowModalError(true);
      setIsDecrypting(false);
    }
    setPostError(null);
  };

  const handleCombine = async () => {
    const combineSuccess = await electionUpdate(
      'combineShares',
      endpoints.editElection(electionID.toString())
    );
    if (combineSuccess && postError === null) {
      setStatus(STATUS.ResultAvailable);
    } else {
      setShowModalError(true);
      setIsOpening(false);
    }
    setPostError(null);
  };

  const getAction = () => {
    switch (status) {
      case STATUS.Initial:
        return (
          <span>
            <OpenButton status={status} handleOpen={handleOpen} />
            <CancelButton status={status} handleCancel={handleCancel} />
          </span>
        );
      case STATUS.Open:
        return (
          <span>
            <CloseButton status={status} handleClose={handleClose} />
            <CancelButton status={status} handleCancel={handleCancel} />
            <VoteButton status={status} electionID={electionID} />
          </span>
        );
      case STATUS.Closed:
        return (
          <span>
            <ShuffleButton
              status={status}
              isShuffling={isShuffling}
              handleShuffle={handleShuffle}
            />
          </span>
        );
      case STATUS.ShuffledBallots:
        return (
          <span>
            <DecryptButton
              status={status}
              isDecrypting={isDecrypting}
              handleDecrypt={handleDecrypt}
            />
          </span>
        );
      case STATUS.DecryptedBallots:
        return (
          <span>
            <CombineButton status={status} handleCombine={handleCombine} />
          </span>
        );
      case STATUS.ResultAvailable:
        return (
          <span>
            <ResultButton status={status} electionID={electionID} />
          </span>
        );
      case STATUS.Canceled:
        return <span> </span>;
      default:
        return <span> --- </span>;
    }
  };
  return { getAction, modalClose, modalCancel };
};

export default useChangeAction;
