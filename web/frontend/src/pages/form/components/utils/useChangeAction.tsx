import React, { useContext, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import * as endpoints from 'components/utils/Endpoints';
import { ID } from 'types/configuration';
import { Action, OngoingAction, Status } from 'types/form';
import { pollForm } from './PollStatus';
import { AuthContext, FlashContext, FlashLevel, ProxyContext } from 'index';
import { useNavigate } from 'react-router';
import { ROUTE_FORM_INDEX } from 'Routes';

import ChooseProxyModal from 'pages/form/components/ChooseProxyModal';
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
import handleLogin from 'pages/session/HandleLogin';
import { UserRole } from 'types/userRole';

const useChangeAction = (
  status: Status,
  formID: ID,
  roster: string[],
  nodeProxyAddresses: Map<string, string>,
  setStatus: (status: Status) => void,
  setResultAvailable: ((available: boolean) => void | null) | undefined,
  setTextModalError: (value: ((prevState: null) => '') | string) => void,
  setShowModalError: (willShow: boolean) => void,
  ongoingAction: OngoingAction,
  setOngoingAction: (action: OngoingAction) => void,
  nodeToSetup: [string, string],
  setNodeToSetup: ([node, proxy]: [string, string]) => void
) => {
  const { t } = useTranslation();
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
  const { isLogged } = useContext(AuthContext);

  const POLLING_INTERVAL = 1000;
  const MAX_ATTEMPTS = 20;

  const modalClose = (
    <ConfirmModal
      showModal={showModalClose}
      setShowModal={setShowModalClose}
      textModal={t('confirmCloseForm')}
      setUserConfirmedAction={setUserConfirmedClosing}
    />
  );
  const modalCancel = (
    <ConfirmModal
      showModal={showModalCancel}
      setShowModal={setShowModalCancel}
      textModal={t('confirmCancelForm')}
      setUserConfirmedAction={setUserConfirmedCanceling}
    />
  );
  const modalDelete = (
    <ConfirmModal
      showModal={showModalDelete}
      setShowModal={setShowModalDelete}
      textModal={t('confirmDeleteForm')}
      setUserConfirmedAction={setUserConfirmedDeleting}
    />
  );

  const modalSetup = (
    <ChooseProxyModal
      roster={roster}
      showModal={showModalProxySetup}
      nodeProxyAddresses={nodeProxyAddresses}
      nodeToSetup={nodeToSetup}
      setNodeToSetup={setNodeToSetup}
      setShowModal={setShowModalProxySetup}
      setUserConfirmedAction={setUserConfirmedProxySetup}
    />
  );

  const formUpdate = async (action: string, endpoint: string) => {
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

  // The previous status is used if there's an error,in which case the form
  // status is set back to this value.
  const pollFormStatus = (previousStatus: Status, nextStatus: Status) => {
    const request = {
      method: 'GET',
      signal: signal,
    };
    // We stop polling when the status has changed to nextStatus
    const match = (s: Status) => s === nextStatus;

    pollForm(
      endpoints.form(pctx.getProxy(), formID),
      request,
      match,
      POLLING_INTERVAL,
      MAX_ATTEMPTS
    )
      .then(
        () => onFullFilled(nextStatus),
        (reason: any) => onRejected(reason, previousStatus)
      )
      .catch((e) => {
        setStatus(previousStatus);
        setGetError(e.message);
      });
  };

  // Start to poll when there is an ongoingAction
  useEffect(() => {
    // use an abortController to stop polling when the component is unmounted

    switch (ongoingAction) {
      case OngoingAction.Initializing:
        // Initializing is handled by each row of the DKG table
        break;
      case OngoingAction.SettingUp:
        // Initializing is handled by each row of the DKG table
        break;
      case OngoingAction.Opening:
        pollFormStatus(Status.Setup, Status.Open);
        break;
      case OngoingAction.Closing:
        pollFormStatus(Status.Open, Status.Closed);
        break;
      case OngoingAction.Canceling:
        pollFormStatus(Status.Open, Status.Canceled);
        break;
      case OngoingAction.Shuffling:
        pollFormStatus(Status.Closed, Status.ShuffledBallots);
        break;
      case OngoingAction.Decrypting:
        pollFormStatus(Status.ShuffledBallots, Status.PubSharesSubmitted);
        break;
      case OngoingAction.Combining:
        pollFormStatus(Status.PubSharesSubmitted, Status.ResultAvailable);
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

        const closeSuccess = await formUpdate(Action.Close, endpoints.editForm(formID));

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

        const cancelSuccess = await formUpdate(Action.Cancel, endpoints.editForm(formID));

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
      const deleteForm = async () => {
        const request = {
          method: 'DELETE',
        };

        const res = await fetch(`/api/evoting/forms/${formID}`, request);
        if (!res.ok) {
          const txt = await res.text();
          fctx.addMessage(`failed to send delete request: ${txt}`, FlashLevel.Error);
          return;
        }

        fctx.addMessage('form deleted', FlashLevel.Info);
        navigate(ROUTE_FORM_INDEX);
      };

      deleteForm();
      setUserConfirmedDeleting(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [userConfirmedDeleting]);

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
          endpoints.editDKGActors(formID),
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
    setOngoingAction(OngoingAction.Initializing);
  };

  const handleSetup = () => {
    setShowModalProxySetup(true);
  };

  const handleOpen = async () => {
    setOngoingAction(OngoingAction.Opening);
    const openSuccess = await formUpdate(Action.Open, endpoints.editForm(formID));

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
    const shuffleSuccess = await formUpdate(Action.Shuffle, endpoints.editShuffle(formID));

    if (!shuffleSuccess) {
      setStatus(Status.Closed);
      setOngoingAction(OngoingAction.None);
    }
  };

  const handleDecrypt = async () => {
    setOngoingAction(OngoingAction.Decrypting);

    const decryptSuccess = await formUpdate(
      Action.BeginDecryption,
      endpoints.editDKGActors(formID)
    );

    if (!decryptSuccess) {
      setStatus(Status.ShuffledBallots);
      setOngoingAction(OngoingAction.None);
    }
  };

  const handleCombine = async () => {
    setOngoingAction(OngoingAction.Combining);
    const combineSuccess = await formUpdate(
      Action.CombineShares,
      endpoints.editForm(formID.toString())
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
    // Except for seeing the results, all actions at least require the users
    // to be logged in
    if (!isLogged && status !== Status.ResultAvailable) {
      return (
        <div>
          {t('notLoggedInActionText1')}
          <button id="login-button" className="text-indigo-600" onClick={() => handleLogin(fctx)}>
            {t('notLoggedInActionText2')}
          </button>
          {t('notLoggedInActionText3')}
        </div>
      );
    }

    // Voters cannot perform any actions except voting and seeing the result
    if ((status < Status.Open || status > Status.Canceled)) {
      return <div>{t('actionTextVoter1')}</div>;
    }

    if (status >= Status.Closed && status < Status.ResultAvailable) {
      return <div>{t('actionTextVoter2')}</div>;
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
            <VoteButton status={status} formID={formID} />
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
            <ResultButton status={status} formID={formID} />
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
