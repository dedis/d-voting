import React, { useContext, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import ConfirmModal from '../modal/ConfirmModal';
import usePostCall from './usePostCall';
import * as endpoints from './Endpoints';
import { ID } from 'types/configuration';
import { Action, OngoingAction, Status } from 'types/election';
import { pollDKG, pollElection } from './PollStatus';
import InitializeButton from 'components/buttons/InitializeButton';
import SetupButton from 'components/buttons/SetupButton';
import OpenButton from 'components/buttons/OpenButton';
import CloseButton from 'components/buttons/CloseButton';
import CancelButton from 'components/buttons/CancelButton';
import VoteButton from 'components/buttons/VoteButton';
import ShuffleButton from 'components/buttons/ShuffleButton';
import DecryptButton from 'components/buttons/DecryptButton';
import CombineButton from 'components/buttons/CombineButton';
import ResultButton from 'components/buttons/ResultButton';
import { NodeStatus } from 'types/node';
import DeleteButton from 'components/buttons/DeleteButton';
import { FlashContext, FlashLevel } from 'index';
import { useNavigate } from 'react-router';
import { ROUTE_ELECTION_INDEX } from 'Routes';

const useChangeAction = (
  status: Status,
  electionID: ID,
  roster: string[],
  nodeProxyAddresses: Map<string, string>,
  setStatus: (status: Status) => void,
  setResultAvailable: ((available: boolean) => void | null) | undefined,
  setTextModalError: (value: ((prevState: null) => '') | string) => void,
  setShowModalError: (willShow: boolean) => void,
  ongoingAction: OngoingAction,
  setOngoingAction: (action: OngoingAction) => void,
  DKGStatuses: Map<string, NodeStatus>,
  setDKGStatuses: (dkgStatuses: Map<string, NodeStatus>) => void
) => {
  const { t } = useTranslation();
  const [isInitializing, setIsInitializing] = useState(false);
  const [isPosting, setIsPosting] = useState(false);
  const [isClosing, setIsClosing] = useState(false);
  const [isCanceling, setIsCanceling] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [showModalClose, setShowModalClose] = useState(false);
  const [showModalCancel, setShowModalCancel] = useState(false);
  const [showModalDelete, setShowModalDelete] = useState(false);
  const [userConfirmedClosing, setUserConfirmedClosing] = useState(false);
  const [userConfirmedCanceling, setUserConfirmedCanceling] = useState(false);
  const [userConfirmedDeleting, setUserConfirmedDeleting] = useState(false);
  const [proxyAddresses, setProxyAddresses] = useState<Map<string, string>>(new Map());
  const [initializedNodes, setInitializedNodes] = useState<Map<string, boolean>>(new Map());
  const [getError, setGetError] = useState(null);
  const [postError, setPostError] = useState(null);
  const sendFetchRequest = usePostCall(setPostError);

  const fctx = useContext(FlashContext);
  const navigate = useNavigate();

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
  const modalDelete = (
    <ConfirmModal
      showModal={showModalDelete}
      setShowModal={setShowModalDelete}
      textModal={t('confirmDeleteElection')}
      setUserConfirmedAction={setUserConfirmedDeleting}
    />
  );

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
    return sendFetchRequest(endpoint, req, setIsPosting);
  };

  const initializeNode = async (address: string) => {
    const request = {
      method: 'POST',
      body: JSON.stringify({
        ElectionID: electionID,
        Proxy: address,
      }),
      headers: {
        'Content-Type': 'application/json',
      },
    };
    return sendFetchRequest(endpoints.dkgActors, request, setIsClosing);
  };

  const onFullFilled = (nextStatus: Status) => {
    if (setGetError !== null && setGetError !== undefined) {
      setGetError(null);
    }

    setStatus(nextStatus);
    setOngoingAction(OngoingAction.None);
  };

  const onRejected = (error: any, previousStatus: Status) => {
    // AbortController sends an AbortError of type DOMException
    // when the component is unmounted, we ignore those
    if (!(error instanceof DOMException)) {
      if (setGetError !== null && setGetError !== undefined) {
        setGetError(error.message);
      }
      setOngoingAction(OngoingAction.None);
      setStatus(previousStatus);
    }
  };

  // The previous status is used if there's an error,in which case the election
  // status is set back to this value.
  const pollElectionStatus = (previousStatus: Status, nextStatus: Status, signal: AbortSignal) => {
    // polling interval
    const interval = 1000;

    const request = {
      method: 'GET',
      signal: signal,
    };
    // We stop polling when the status has changed
    const match = (s: Status) => s !== previousStatus;

    pollElection(endpoints.election(electionID), request, match, interval)
      .then(
        () => onFullFilled(nextStatus),
        (reason: any) => onRejected(reason, previousStatus)
      )
      .catch((e) => {
        setStatus(previousStatus);
        setGetError(e.message);
      });
  };

  const pollDKGStatus = (proxy: string, statusToMatch: NodeStatus, signal: AbortSignal) => {
    const interval = 1000;

    const request = {
      method: 'GET',
      signal: signal,
    };

    const match = (s: NodeStatus) => s === statusToMatch;

    return pollDKG(endpoints.getDKGActors(proxy, electionID), request, match, interval);
  };

  // Start to poll when there is an ongoingAction
  useEffect(() => {
    // use an abortController to stop polling when the component is unmounted
    const abortController = new AbortController();
    const signal = abortController.signal;

    switch (ongoingAction) {
      case OngoingAction.Initializing:
        // Initialize each of the node participating in the election
        const promises: Promise<unknown>[] = Array.from(nodeProxyAddresses.values()).map(
          (proxy) => {
            return pollDKGStatus(proxy, NodeStatus.Initialized, signal);
          }
        );

        Promise.all(promises).then(
          () => {
            onFullFilled(Status.Initialized);
            const newDKGStatuses = new Map(DKGStatuses);
            nodeProxyAddresses.forEach((_proxy, node) =>
              newDKGStatuses.set(node, NodeStatus.Initialized)
            );
            setDKGStatuses(newDKGStatuses);
          },
          (reason: any) => onRejected(reason, Status.Initial)
        );

        break;
      case OngoingAction.SettingUp:
        // Setup the first node in the roster
        const node = roster[0];
        pollDKGStatus(nodeProxyAddresses.get(node), NodeStatus.Setup, signal)
          .then(
            () => {
              onFullFilled(Status.Setup);
              const newDKGStatuses = new Map(DKGStatuses);
              newDKGStatuses.set(node, NodeStatus.Setup);
              setDKGStatuses(newDKGStatuses);
            },
            (reason: any) => {
              onRejected(reason, Status.Initialized);
              const newDKGStatuses = new Map(DKGStatuses);
              newDKGStatuses.set(node, NodeStatus.Failed);
              setDKGStatuses(newDKGStatuses);
            }
          )
          .catch((e) => {
            setStatus(Status.Initialized);
            setGetError(e.message);
            setShowModalError(true);
          });
        break;
      case OngoingAction.Opening:
        pollElectionStatus(Status.Setup, Status.Open, signal);
        break;
      case OngoingAction.Closing:
        pollElectionStatus(Status.Open, Status.Closed, signal);
        break;
      case OngoingAction.Canceling:
        pollElectionStatus(Status.Open, Status.Canceled, signal);
        break;
      case OngoingAction.Shuffling:
        pollElectionStatus(Status.Closed, Status.ShuffledBallots, signal);
        break;
      case OngoingAction.Decrypting:
        pollElectionStatus(Status.ShuffledBallots, Status.PubSharesSubmitted, signal);
        break;
      case OngoingAction.Combining:
        pollElectionStatus(Status.PubSharesSubmitted, Status.ResultAvailable, signal);
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
      setShowModalError(true);
      setPostError(null);
    }
  }, [postError]);

  useEffect(() => {
    if (getError !== null) {
      setTextModalError(getError);
      setShowModalError(true);
      setGetError(null);
    }
  }, [getError]);

  useEffect(() => {
    //check if close button was clicked and the user validated the confirmation window
    if (isClosing && userConfirmedClosing) {
      const close = async () => {
        const closeSuccess = await electionUpdate(Action.Close, endpoints.editElection(electionID));
        if (closeSuccess) {
          setOngoingAction(OngoingAction.Closing);
        }

        setUserConfirmedClosing(false);
      };

      close();
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
        }
        setUserConfirmedCanceling(false);
      };

      cancel();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isCanceling, sendFetchRequest, setShowModalError, setStatus, userConfirmedCanceling]);

  useEffect(() => {
    if (isDeleting && userConfirmedDeleting) {
      const deleteElection = async () => {
        const request = {
          method: 'DELETE',
        };

        const res = await fetch(`/api/evoting/elections/${electionID}`, request);
        if (!res.ok) {
          const txt = await res.text();
          fctx.addMessage(`failed to send delete request: ${txt}`, FlashLevel.Error);
          return;
        }

        fctx.addMessage('election deleted', FlashLevel.Info);
        navigate(ROUTE_ELECTION_INDEX);
      };

      deleteElection();
      setIsDeleting(false);
      setUserConfirmedDeleting(false);
    }
  });

  useEffect(() => {
    if (isInitializing) {
      const initialize = async () => {
        proxyAddresses.forEach(async (address) => {
          const initSuccess = await initializeNode(address);

          if (initSuccess) {
            const initNodes = new Map(initializedNodes);
            initNodes.set(address, true);
            setInitializedNodes(initNodes);

            // All post request to initialize the nodes have been sent
            if (!Array.from(initializedNodes.values()).includes(false)) {
              setIsInitializing(false);
              setOngoingAction(OngoingAction.Initializing);
            }
          }
        });
      };

      initialize();
    }
  }, [isInitializing]);

  const handleInitialize = () => {
    // initialize the address of the proxies with the address of the node
    if (proxyAddresses.size === 0) {
      const initProxAddresses = new Map(proxyAddresses);
      roster.forEach((node) => initProxAddresses.set(node, node));
      setProxyAddresses(initProxAddresses);
    }
    setIsInitializing(true);
  };

  const handleSetup = async () => {
    const setupSuccess = await electionUpdate(Action.Setup, endpoints.editDKGActors(electionID));

    if (setupSuccess) {
      setOngoingAction(OngoingAction.SettingUp);
    }
  };

  const handleOpen = async () => {
    const openSuccess = await electionUpdate(Action.Open, endpoints.editElection(electionID));
    if (openSuccess) {
      setOngoingAction(OngoingAction.Opening);
    }
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
    if (shuffleSuccess) {
      setOngoingAction(OngoingAction.Shuffling);
    }
  };

  const handleDecrypt = async () => {
    const decryptSuccess = await electionUpdate(
      Action.BeginDecryption,
      endpoints.editDKGActors(electionID)
    );
    if (decryptSuccess) {
      setOngoingAction(OngoingAction.Decrypting);
    }
  };

  const handleCombine = async () => {
    const combineSuccess = await electionUpdate(
      Action.CombineShares,
      endpoints.editElection(electionID.toString())
    );
    if (combineSuccess && postError === null) {
      setOngoingAction(OngoingAction.Combining);
    }
  };

  const handleDelete = () => {
    setShowModalDelete(true);
    setIsDeleting(true);
  };

  const getAction = () => {
    switch (status) {
      case Status.Initial:
        return (
          <>
            <InitializeButton
              status={status}
              handleInitialize={handleInitialize}
              ongoingAction={ongoingAction}
            />
            <DeleteButton handleDelete={handleDelete} />
          </>
        );
      case Status.Initialized:
        return (
          <>
            <SetupButton status={status} handleSetup={handleSetup} ongoingAction={ongoingAction} />
            <DeleteButton handleDelete={handleDelete} />
          </>
        );
      case Status.Setup:
        return (
          <>
            <OpenButton status={status} handleOpen={handleOpen} ongoingAction={ongoingAction} />
            <DeleteButton handleDelete={handleDelete} />
          </>
        );
      case Status.Open:
        return (
          <>
            <CloseButton status={status} handleClose={handleClose} ongoingAction={ongoingAction} />
            <CancelButton
              status={status}
              handleCancel={handleCancel}
              ongoingAction={ongoingAction}
            />
            <VoteButton status={status} electionID={electionID} />
            <DeleteButton handleDelete={handleDelete} />
          </>
        );
      case Status.Closed:
        return (
          <>
            <ShuffleButton
              status={status}
              handleShuffle={handleShuffle}
              ongoingAction={ongoingAction}
            />
            <DeleteButton handleDelete={handleDelete} />
          </>
        );
      case Status.ShuffledBallots:
        return (
          <>
            <DecryptButton
              status={status}
              handleDecrypt={handleDecrypt}
              ongoingAction={ongoingAction}
            />
            <DeleteButton handleDelete={handleDelete} />
          </>
        );
      case Status.PubSharesSubmitted:
        return (
          <>
            <CombineButton
              status={status}
              handleCombine={handleCombine}
              ongoingAction={ongoingAction}
            />
            <DeleteButton handleDelete={handleDelete} />
          </>
        );
      case Status.ResultAvailable:
        return (
          <>
            <ResultButton status={status} electionID={electionID} />
            <DeleteButton handleDelete={handleDelete} />
          </>
        );
      case Status.Canceled:
        return (
          <>
            <DeleteButton handleDelete={handleDelete} />
          </>
        );
      default:
        return (
          <>
            <DeleteButton handleDelete={handleDelete} />
          </>
        );
    }
  };
  return { getAction, modalClose, modalCancel, modalDelete };
};

export default useChangeAction;
