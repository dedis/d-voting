import React, { useContext, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import * as endpoints from 'components/utils/Endpoints';
import { ID } from 'types/configuration';
import { Action, OngoingAction, Status } from 'types/election';
import { pollDKG, pollElection } from './PollStatus';
import { NodeStatus } from 'types/node';
import { FlashContext, FlashLevel, ProxyContext } from 'index';
import { useNavigate } from 'react-router';
import { ROUTE_ELECTION_INDEX } from 'Routes';

import ChooseProxyModal from 'components/modal/ChooseProxyModal';
import ConfirmModal from 'components/modal/ConfirmModal';
import usePostCall from 'components/utils/usePostCall';
import InitializeButton from '../ActionButtons/InitializeButton';
import DeleteButton from '../ActionButtons/DeleteButton';
import SetupButton from '../ActionButtons/SetupButton';
import CancelButton from '../ActionButtons/CancelButton';
import CloseButton from '../ActionButtons/CloseButton';
import CombineButton from '../ActionButtons/CombineButton';
import DecryptButton from '../ActionButtons/DecryptButton';
import OpenButton from '../ActionButtons/OpenButton';
import ResultButton from '../ActionButtons/ResultButton';
import ShuffleButton from '../ActionButtons/ShuffleButton';
import VoteButton from '../ActionButtons/VoteButton';
import NoActionAvailable from '../ActionButtons/NoActionAvailable';

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
  nodeToSetup: [string, string],
  setNodeToSetup: ([node, proxy]: [string, string]) => void,
  DKGStatuses: Map<string, NodeStatus>,
  setDKGStatuses: (dkgStatuses: Map<string, NodeStatus>) => void
) => {
  const { t } = useTranslation();
  const [isInitializing, setIsInitializing] = useState(false);
  const [, setIsPosting] = useState(false);

  const [showModalProxySetup, setShowModalProxySetup] = useState(false);
  const [showModalClose, setShowModalClose] = useState(false);
  const [showModalCancel, setShowModalCancel] = useState(false);
  const [showModalDelete, setShowModalDelete] = useState(false);

  const [userConfirmedProxySetup, setUserConfirmedProxySetup] = useState(false);
  const [userConfirmedClosing, setUserConfirmedClosing] = useState(false);
  const [userConfirmedCanceling, setUserConfirmedCanceling] = useState(false);
  const [userConfirmedDeleting, setUserConfirmedDeleting] = useState(false);

  const [getError, setGetError] = useState(null);
  const [postError, setPostError] = useState(null);
  const sendFetchRequest = usePostCall(setPostError);
  const abortController = new AbortController();
  const signal = abortController.signal;

  const fctx = useContext(FlashContext);
  const navigate = useNavigate();
  const pctx = useContext(ProxyContext);

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

  const modalSetup = (
    <ChooseProxyModal
      showModal={showModalProxySetup}
      nodeProxyAddresses={nodeProxyAddresses}
      nodeToSetup={nodeToSetup}
      setNodeToSetup={setNodeToSetup}
      setShowModal={setShowModalProxySetup}
      setUserConfirmedAction={setUserConfirmedProxySetup}
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

  const initializeNode = async (proxy: string) => {
    const request = {
      method: 'POST',
      body: JSON.stringify({
        ElectionID: electionID,
        Proxy: proxy,
      }),
      headers: {
        'Content-Type': 'application/json',
      },
    };
    return sendFetchRequest(endpoints.dkgActors, request, setIsPosting);
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
  const pollElectionStatus = (previousStatus: Status, nextStatus: Status) => {
    // polling interval
    const interval = 1000;

    const request = {
      method: 'GET',
      signal: signal,
    };
    // We stop polling when the status has changed to nextStatus
    const match = (s: Status) => s === nextStatus;

    pollElection(endpoints.election(pctx.getProxy(), electionID), request, match, interval)
      .then(
        () => onFullFilled(nextStatus),
        (reason: any) => onRejected(reason, previousStatus)
      )
      .catch((e) => {
        setStatus(previousStatus);
        setGetError(e.message);
      });
  };

  const pollDKGStatus = (proxy: string, statusToMatch: NodeStatus) => {
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

    switch (ongoingAction) {
      case OngoingAction.Initializing:
        if (nodeProxyAddresses !== null) {
          const promises = Array.from(nodeProxyAddresses.values()).map((proxy) => {
            return pollDKGStatus(proxy, NodeStatus.Initialized);
          });

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
        }
        break;
      case OngoingAction.SettingUp:
        if (nodeToSetup !== null) {
          pollDKGStatus(nodeToSetup[1], NodeStatus.Setup)
            .then(
              () => {
                onFullFilled(Status.Setup);
                const newDKGStatuses = new Map(DKGStatuses);
                newDKGStatuses.set(nodeToSetup[0], NodeStatus.Setup);
                setDKGStatuses(newDKGStatuses);
              },
              (reason: any) => {
                onRejected(reason, Status.Initialized);

                if (!(reason instanceof DOMException)) {
                  const newDKGStatuses = new Map(DKGStatuses);
                  newDKGStatuses.set(nodeToSetup[0], NodeStatus.Failed);
                  setDKGStatuses(newDKGStatuses);
                }
              }
            )
            .catch((e) => {
              setStatus(Status.Initialized);
              setGetError(e.message);
              setShowModalError(true);
            });
        }
        break;
      case OngoingAction.Opening:
        pollElectionStatus(Status.Setup, Status.Open);
        break;
      case OngoingAction.Closing:
        pollElectionStatus(Status.Open, Status.Closed);
        break;
      case OngoingAction.Canceling:
        pollElectionStatus(Status.Open, Status.Canceled);
        break;
      case OngoingAction.Shuffling:
        pollElectionStatus(Status.Closed, Status.ShuffledBallots);
        break;
      case OngoingAction.Decrypting:
        pollElectionStatus(Status.ShuffledBallots, Status.PubSharesSubmitted);
        break;
      case OngoingAction.Combining:
        pollElectionStatus(Status.PubSharesSubmitted, Status.ResultAvailable);
        setResultAvailable(true);
        break;
      default:
        break;
    }

    return () => {
      abortController.abort();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ongoingAction, nodeProxyAddresses]);

  useEffect(() => {
    if (postError !== null) {
      setTextModalError(t('errorAction', { error: postError }));
      setShowModalError(true);
      setPostError(null);
      abortController.abort();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [postError]);

  useEffect(() => {
    if (getError !== null) {
      setTextModalError(t('errorAction', { error: getError }));
      setShowModalError(true);
      setGetError(null);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [getError]);

  useEffect(() => {
    //check if close button was clicked and the user validated the confirmation window
    if (userConfirmedClosing) {
      const close = async () => {
        setOngoingAction(OngoingAction.Closing);

        const closeSuccess = await electionUpdate(Action.Close, endpoints.editElection(electionID));

        if (!closeSuccess) {
          setStatus(Status.Open);
          setOngoingAction(OngoingAction.None);
        }

        setUserConfirmedClosing(false);
      };

      close();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [userConfirmedClosing]);

  useEffect(() => {
    if (userConfirmedCanceling) {
      const cancel = async () => {
        setOngoingAction(OngoingAction.Canceling);

        const cancelSuccess = await electionUpdate(
          Action.Cancel,
          endpoints.editElection(electionID)
        );

        if (!cancelSuccess) {
          setStatus(Status.Open);
          setOngoingAction(OngoingAction.None);
        }
        setUserConfirmedCanceling(false);
      };

      cancel();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [userConfirmedCanceling]);

  useEffect(() => {
    if (userConfirmedDeleting) {
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
      setUserConfirmedDeleting(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [userConfirmedDeleting]);

  useEffect(() => {
    if (isInitializing) {
      const initialize = async (proxy: string) => {
        setOngoingAction(OngoingAction.Initializing);
        const initSuccess = await initializeNode(proxy);

        if (!initSuccess) {
          setStatus(Status.Initial);
          setOngoingAction(OngoingAction.None);
        }
      };

      const promises = Array.from(nodeProxyAddresses.values()).map((proxy) => {
        return initialize(proxy);
      });

      Promise.all(promises).then(() => {
        setIsInitializing(false);
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isInitializing]);

  useEffect(() => {
    if (userConfirmedProxySetup) {
      const setup = async () => {
        setOngoingAction(OngoingAction.SettingUp);

        const request = {
          method: 'PUT',
          body: JSON.stringify({
            Action: Action.Setup,
            Proxy: nodeToSetup[1],
          }),
          headers: {
            'Content-Type': 'application/json',
          },
        };

        const setupSuccess = await sendFetchRequest(
          endpoints.editDKGActors(electionID),
          request,
          setIsPosting
        );

        if (!setupSuccess) {
          setStatus(Status.Initialized);
          setOngoingAction(OngoingAction.None);
        }
        setUserConfirmedProxySetup(false);
      };

      setup();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [userConfirmedProxySetup]);

  const handleInitialize = () => {
    setIsInitializing(true);
  };

  const handleSetup = () => {
    setShowModalProxySetup(true);
  };

  const handleOpen = async () => {
    setOngoingAction(OngoingAction.Opening);
    const openSuccess = await electionUpdate(Action.Open, endpoints.editElection(electionID));

    if (!openSuccess) {
      setStatus(Status.Setup);
      setOngoingAction(OngoingAction.None);
    }
  };

  const handleClose = () => {
    setShowModalClose(true);
  };

  const handleCancel = () => {
    setShowModalCancel(true);
  };

  const handleShuffle = async () => {
    setOngoingAction(OngoingAction.Shuffling);
    const shuffleSuccess = await electionUpdate(Action.Shuffle, endpoints.editShuffle(electionID));

    if (!shuffleSuccess) {
      setStatus(Status.Closed);
      setOngoingAction(OngoingAction.None);
    }
  };

  const handleDecrypt = async () => {
    setOngoingAction(OngoingAction.Decrypting);

    const decryptSuccess = await electionUpdate(
      Action.BeginDecryption,
      endpoints.editDKGActors(electionID)
    );

    if (!decryptSuccess) {
      setStatus(Status.ShuffledBallots);
      setOngoingAction(OngoingAction.None);
    }
  };

  const handleCombine = async () => {
    setOngoingAction(OngoingAction.Combining);
    const combineSuccess = await electionUpdate(
      Action.CombineShares,
      endpoints.editElection(electionID.toString())
    );

    if (!combineSuccess) {
      setStatus(Status.PubSharesSubmitted);
      setOngoingAction(OngoingAction.None);
    }
  };

  const handleDelete = () => {
    setShowModalDelete(true);
  };

  const getAction = () => {
    if (Array.from(DKGStatuses.values()).includes(NodeStatus.Unreachable)) {
      return (
        <>
          <NoActionAvailable />
          <DeleteButton handleDelete={handleDelete} />
        </>
      );
    }

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
      default:
        return (
          <>
            <DeleteButton handleDelete={handleDelete} />
          </>
        );
    }
  };
  return { getAction, modalClose, modalCancel, modalDelete, modalSetup };
};

export default useChangeAction;
