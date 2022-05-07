import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import ConfirmModal from '../modal/ConfirmModal';
import usePostCall from './usePostCall';
import * as endpoints from './Endpoints';
import { ID } from 'types/configuration';
import { Action, NodeStatus, OngoingAction, Status } from 'types/election';
import {
  CancelButton,
  CloseButton,
  CombineButton,
  DecryptButton,
  InitializeButton,
  NoActionAvailable,
  OpenButton,
  ResultButton,
  SetupButton,
  ShuffleButton,
  VoteButton,
} from './ActionButtons';
import { poll } from './usePolling';
import AddProxyAddressesModal from 'components/modal/AddProxyAddressesModal';

const useChangeAction = (
  status: Status,
  electionID: ID,
  roster: string[],
  setStatus: (status: Status) => void,
  setResultAvailable: ((available: boolean) => void | null) | undefined,
  setTextModalError: (value: ((prevState: null) => '') | string) => void,
  setShowModalError: (willShow: boolean) => void,
  setGetError: (error: string) => void,
  setHasInitialized: (init: boolean) => void,
  ongoingAction: OngoingAction,
  setOngoingAction: (action: OngoingAction) => void
) => {
  const { t } = useTranslation();
  const [isInitializing, setIsInitializing] = useState(false);
  //const [hasInitialized, setHasInitialized] = useState(false);
  const [isClosing, setIsClosing] = useState(false);
  const [isCanceling, setIsCanceling] = useState(false);
  const [showModalClose, setShowModalClose] = useState(false);
  const [showModalCancel, setShowModalCancel] = useState(false);
  const [showModalAddProxy, setShowModalAddProxy] = useState(false);
  const [userConfirmedClosing, setUserConfirmedClosing] = useState(false);
  const [userConfirmedCanceling, setUserConfirmedCanceling] = useState(false);
  const [userConfirmedAddProxy, setUserConfirmedAddProx] = useState(false);
  const [proxyAddresses, setProxyAddresses] = useState<Map<string, string>>(new Map());
  const [initializedNodes, setInitializedNodes] = useState<Map<string, boolean>>(new Map());
  //const [ongoingAction, setOngoingAction] = useState(OngoingAction.None);

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
  const pollStatus = (
    endpoint: string,
    statusToMatch: Status | NodeStatus,
    previousStatus: Status,
    nextStatus: Status,
    signal: AbortSignal
  ) => {
    const interval = 1000;
    const request = {
      method: 'GET',
      signal: signal,
    };

    const onFullFilled = () => {
      if (setGetError !== null && setGetError !== undefined) {
        setGetError(null);
      }

      setStatus(nextStatus);
      setOngoingAction(OngoingAction.None);
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

    const match = (s: Status | NodeStatus) => s === statusToMatch;

    poll(endpoint, request, match, interval)
      .then(onFullFilled, onRejected)
      .catch((e) => {
        setStatus(previousStatus);
        setGetError(e.message);
        setShowModalError(true);
      });
  };

  // TODO poll on every Status
  useEffect(() => {
    // use an abortController to stop polling when the component is unmounted
    const abortController = new AbortController();
    const signal = abortController.signal;

    switch (ongoingAction) {
      case OngoingAction.Initializing:
        pollStatus(
          endpoints.editDKGActors(electionID),
          NodeStatus.Initialized,
          Status.Initial,
          Status.Initialized,
          signal
        );
        break;
      case OngoingAction.SettingUp:
        pollStatus(
          endpoints.editDKGActors(electionID),
          NodeStatus.Setup,
          Status.Initialized,
          Status.Setup,
          signal
        );
        break;
      case OngoingAction.Opening:
        pollStatus(
          endpoints.election(electionID),
          Status.Open,
          // TODO: works only with the mock !
          // Will go back to Status.Initial using the real backend
          Status.Setup,
          Status.Open,
          signal
        );
        break;
      case OngoingAction.Closing:
        pollStatus(
          endpoints.election(electionID),
          Status.Closed,
          Status.Open,
          Status.Closed,
          signal
        );
        break;
      case OngoingAction.Canceling:
        pollStatus(
          endpoints.election(electionID),
          Status.Canceled,
          Status.Open,
          Status.Canceled,
          signal
        );
        break;
      case OngoingAction.Shuffling:
        pollStatus(
          endpoints.election(electionID),
          Status.ShuffledBallots,
          Status.Closed,
          Status.ShuffledBallots,
          signal
        );
        break;
      case OngoingAction.Decrypting:
        pollStatus(
          endpoints.election(electionID),
          Status.PubSharesSubmitted,
          Status.ShuffledBallots,
          Status.PubSharesSubmitted,
          signal
        );
        break;
      case OngoingAction.Combining:
        pollStatus(
          endpoints.election(electionID),
          Status.ResultAvailable,
          Status.PubSharesSubmitted,
          Status.ResultAvailable,
          signal
        );
        setResultAvailable(true);
        break;
      default:
        break;
    }

    return () => {
      abortController.abort();
    };
  }, [ongoingAction]);

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
        const closeSuccess = await electionUpdate(Action.Close, endpoints.editElection(electionID));
        if (closeSuccess) {
          setOngoingAction(OngoingAction.Closing);
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
          Action.Cancel,
          endpoints.editElection(electionID)
        );
        if (cancelSuccess) {
          setOngoingAction(OngoingAction.Canceling);
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
              setIsInitializing(false);

              // TODO This should persist
              setHasInitialized(true);
              setOngoingAction(OngoingAction.Initializing);
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

  // Setup one of the node and then start polling to know when all the nodes
  // have been setup
  const handleSetup = async () => {
    const setupSuccess = await electionUpdate(Action.Setup, endpoints.editDKGActors(electionID));

    if (setupSuccess && postError === null) {
      setOngoingAction(OngoingAction.SettingUp);
    } else {
      setShowModalError(true);
    }
    setPostError(null);
  };

  const handleOpen = async () => {
    const openSuccess = await electionUpdate(Action.Open, endpoints.editElection(electionID));
    if (openSuccess && postError === null) {
      setOngoingAction(OngoingAction.Opening);
    } else {
      setShowModalError(true);
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
    const shuffleSuccess = await electionUpdate(Action.Shuffle, endpoints.editShuffle(electionID));
    if (shuffleSuccess && postError === null) {
      setOngoingAction(OngoingAction.Shuffling);
    } else {
      setShowModalError(true);
    }
    setPostError(null);
  };

  const handleDecrypt = async () => {
    const decryptSuccess = await electionUpdate(
      Action.BeginDecryption,
      endpoints.editDKGActors(electionID)
    );
    if (decryptSuccess && postError === null) {
      setOngoingAction(OngoingAction.Decrypting);
    } else {
      setShowModalError(true);
      //setIsDecrypting(false);
    }
    setPostError(null);
  };

  const handleCombine = async () => {
    const combineSuccess = await electionUpdate(
      Action.CombineShares,
      endpoints.editElection(electionID.toString())
    );
    if (combineSuccess && postError === null) {
      setOngoingAction(OngoingAction.Combining);
    } else {
      setShowModalError(true);
    }
    setPostError(null);
  };

  const getAction = () => {
    console.log(ongoingAction);
    switch (status) {
      case Status.Initial:
        return (
          <span>
            <InitializeButton
              status={status}
              handleInitialize={handleInitialize}
              ongoingAction={ongoingAction}
            />
          </span>
        );
      case Status.Initialized:
        return (
          <span>
            <SetupButton status={status} handleSetup={handleSetup} ongoingAction={ongoingAction} />
          </span>
        );
      case Status.Setup:
        return (
          <span>
            <OpenButton status={status} handleOpen={handleOpen} ongoingAction={ongoingAction} />
          </span>
        );
      case Status.Open:
        return (
          <span>
            <CloseButton status={status} handleClose={handleClose} ongoingAction={ongoingAction} />
            <CancelButton
              status={status}
              handleCancel={handleCancel}
              ongoingAction={ongoingAction}
            />
            <VoteButton status={status} electionID={electionID} />
          </span>
        );
      case Status.Closed:
        return (
          <span>
            <ShuffleButton
              status={status}
              handleShuffle={handleShuffle}
              ongoingAction={ongoingAction}
            />
          </span>
        );
      case Status.ShuffledBallots:
        return (
          <span>
            <DecryptButton
              status={status}
              handleDecrypt={handleDecrypt}
              ongoingAction={ongoingAction}
            />
          </span>
        );
      case Status.PubSharesSubmitted:
        return (
          <CombineButton
            status={status}
            handleCombine={handleCombine}
            ongoingAction={ongoingAction}
          />
        );
      case Status.ResultAvailable:
        return (
          <span>
            <ResultButton status={status} electionID={electionID} />
          </span>
        );
      case Status.Canceled:
        return (
          <span>
            <NoActionAvailable />
          </span>
        );
      default:
        return (
          <span>
            <NoActionAvailable />
          </span>
        );
    }
  };
  return { getAction, modalClose, modalCancel, modalAddProxyAddresses };
};

export default useChangeAction;
