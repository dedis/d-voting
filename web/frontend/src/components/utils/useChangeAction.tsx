import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import ConfirmModal from '../modal/ConfirmModal';
import usePostCall from './usePostCall';
import * as endpoints from './Endpoints';
import { ID } from 'types/configuration';
import { ACTION, STATUS } from 'types/election';
import {
  CancelButton,
  CloseButton,
  CombineSharesButton,
  DecryptButton,
  InitializeButton,
  OpenButton,
  ResultButton,
  SetupButton,
  ShuffleButton,
} from './ActionButtons';
import { poll } from './usePolling';
import AddProxyAddressesModal from 'components/modal/AddProxyAddressesModal';

const useChangeAction = (
  status: STATUS,
  electionID: ID,
  roster: string[],
  setStatus: (status: STATUS) => void,
  setResultAvailable: ((available: boolean) => void | null) | undefined,
  setTextModalError: (value: ((prevState: null) => '') | string) => void,
  setShowModalError: (willShow: boolean) => void,
  setGetError: (error: string) => void
) => {
  const { t } = useTranslation();
  const [isInitializing, setIsInitializing] = useState(false);
  const [hasInitialized, setHasInitialized] = useState(false);
  const [isSettingUp, setIsSettingUp] = useState(false);
  const [isOpening, setIsOpening] = useState(false);
  const [isClosing, setIsClosing] = useState(false);
  const [isCanceling, setIsCanceling] = useState(false);
  const [isShuffling, setIsShuffling] = useState(false);
  const [isDecrypting, setIsDecrypting] = useState(false);
  const [isCombining, setIsCombining] = useState(false);
  const [showModalOpen, setShowModalOpen] = useState(false);
  const [showModalClose, setShowModalClose] = useState(false);
  const [showModalCancel, setShowModalCancel] = useState(false);
  const [showModalAddProxy, setShowModalAddProxy] = useState(false);
  const [userConfirmedClosing, setUserConfirmedClosing] = useState(false);
  const [userConfirmedCanceling, setUserConfirmedCanceling] = useState(false);
  const [userConfirmedAddProxy, setUserConfirmedAddProx] = useState(false);
  const [proxyAddresses, setProxyAddresses] = useState<Map<string, string>>(new Map());
  const [initializedNodes, setInitializedNodes] = useState<Map<string, boolean>>(new Map());

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

  const modalAddProxyAddresses = (
    <AddProxyAddressesModal
      roster={roster}
      proxyAddresses={proxyAddresses}
      setProxyAddresses={setProxyAddresses}
      showModal={showModalAddProxy}
      setShowModal={setShowModalAddProxy}
      setUserConfirmedAction={setUserConfirmedAddProx}
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

  const initializeNode = async (address: string) => {
    const request = {
      method: 'POST',
      body: JSON.stringify({
        ElectionID: electionID,
        ProxyAddress: address,
      }),
      headers: {
        'Content-Type': 'application/json',
      },
    };
    return sendFetchRequest(endpoints.dkgActors, request, setIsClosing);
  };

  // Start to poll on the given endpoint, statusToMatch is the status we are
  // waiting for to stop polling. The previous status is used if there's an error,
  // in which case the election status is set back to this value.
  const pollStatus = (endpoint: string, statusToMatch: STATUS, previousStatus: STATUS, signal) => {
    const interval = 1000;
    const request = {
      method: 'GET',
      signal: signal,
    };

    const onFullFilled = () => {
      if (setGetError !== null && setGetError !== undefined) {
        setGetError(null);
      }

      setStatus(statusToMatch);
    };

    const onRejected = (error) => {
      // AbortController sends an AbortError of type DOMException
      // when the component is unmounted, we ignore those
      if (!(error instanceof DOMException)) {
        if (setGetError !== null && setGetError !== undefined) {
          setGetError(error.message);
        }

        setStatus(previousStatus);
      }
    };

    const match = (s: STATUS) => s === statusToMatch;

    poll(endpoint, request, match, interval)
      .then(onFullFilled, onRejected)
      .catch((e) => {
        setStatus(previousStatus);
        setGetError(e.message);
        setShowModalError(true);
      });
  };

  useEffect(() => {
    // use an abortController to stop polling when the component is unmounted
    const abortController = new AbortController();
    const signal = abortController.signal;

    return () => {
      abortController.abort();
    };
  }, [status]);

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
        const closeSuccess = await electionUpdate(ACTION.Close, endpoints.editElection(electionID));
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
          ACTION.Cancel,
          endpoints.editElection(electionID)
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

  useEffect(() => {
    if (isInitializing && userConfirmedAddProxy) {
      const initialize = async () => {
        proxyAddresses.forEach(async (address) => {
          const initSuccess = await initializeNode(address);

          if (initSuccess && postError == null) {
            const initNodes = new Map(initializedNodes);
            initNodes.set(address, true);
            setInitializedNodes(initNodes);

            // All the nodes have been initialized
            if (!Array.from(initializedNodes.values()).includes(false)) {
              // TODO poll to be sure
              setStatus(STATUS.Initialized);
              // TODO This should persist
              setHasInitialized(true);
            }
          } else {
            setShowModalError(true);
          }
          setPostError(null);
        });
      };

      initialize();
    }
  }, [isInitializing, userConfirmedAddProxy]);

  const handleInitialize = () => {
    console.log(proxyAddresses);
    // initialize state
    if (proxyAddresses.size === 0) {
      const initProxAddresses = new Map(proxyAddresses);
      roster.forEach((node) => initProxAddresses.set(node, ''));
      setProxyAddresses(initProxAddresses);
    }

    setShowModalAddProxy(true);
    setIsInitializing(true);
  };

  const handleSetup = async () => {};

  const handleOpen = async () => {
    const openSuccess = await electionUpdate(ACTION.Open, endpoints.editElection(electionID));
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
    const shuffleSuccess = await electionUpdate(ACTION.Shuffle, endpoints.editShuffle(electionID));
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
      ACTION.BeginDecryption,
      endpoints.editDKGActors(electionID)
    );
    if (decryptSuccess && postError === null) {
      // TODO : setResultAvailable is undefined when the decryption is clicked
      if (setResultAvailable !== null && setResultAvailable !== undefined) {
        setResultAvailable(true);
      }
      setStatus(STATUS.ResultAvailable);
    } else {
      setShowModalError(true);
      setIsDecrypting(false);
    }
    setPostError(null);
  };

  const handleCombineShares = () => {};

  const getAction = () => {
    switch (status) {
      case STATUS.Initial:
        return (
          <span>
            <InitializeButton status={status} handleInitialize={handleInitialize} />
          </span>
        );
      case STATUS.Initialized:
        return (
          <span>
            <SetupButton status={status} handleSetup={handleSetup} />
          </span>
        );
      case STATUS.OnGoingSetup:
        return <span>{t('statusOnGoingSetup')}</span>;
      case STATUS.Setup:
        return (
          <span>
            <OpenButton status={status} handleOpen={handleOpen} />
          </span>
        );
      case STATUS.Open:
        return (
          <span>
            <CloseButton status={status} handleClose={handleClose} />
            <CancelButton status={status} handleCancel={handleCancel} />
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
      case STATUS.OnGoingShuffle:
        return <span>{t('statusOnGoingShuffle')}</span>;

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
      case STATUS.OnGoingDecryption:
        return <span>{t('statusOnGoingDecryption')}</span>;
      case STATUS.PubSharesSubmitted:
        return <CombineSharesButton status={status} handleCombineShares={handleCombineShares} />;
      case STATUS.ResultAvailable:
        return (
          <span>
            <ResultButton status={status} electionID={electionID} />
          </span>
        );
      case STATUS.Canceled:
        return <span> --- </span>;
      default:
        return <span> --- </span>;
    }
  };
  return { getAction, modalClose, modalCancel, modalAddProxyAddresses };
};

export default useChangeAction;
