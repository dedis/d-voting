import React, { FC, useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

import useForm from 'components/utils/useForm';
import { OngoingAction, Status } from 'types/form';
import Modal from 'components/modal/Modal';
import { ROUTE_FORM_INDEX } from '../../Routes';
import StatusTimeline from './components/StatusTimeline';
import Loading from 'pages/Loading';
import Action from './components/Action';
import { InternalDKGInfo, NodeStatus } from 'types/node';
import useGetResults from './components/utils/useGetResults';
import UserIDTable from './components/UserIDTable';
import DKGStatusTable from './components/DKGStatusTable';
import LoadingButton from './components/LoadingButton';
import { default as i18n } from 'i18next';

const FormShow: FC = () => {
  const { t } = useTranslation();
  const { formId } = useParams();
  const {
    loading,
    formID,
    status,
    setStatus,
    roster,
    setResult,
    configObj,
    setIsResultSet,
    voters,
    error,
  } = useForm(formId);

  const [, setError] = useState(null);
  const [isResultAvailable, setIsResultAvailable] = useState(false);
  const { getResults } = useGetResults();

  const [textModalError, setTextModalError] = useState(null);
  const [showModalError, setShowModalError] = useState(false);

  const [ongoingAction, setOngoingAction] = useState(OngoingAction.None);

  const [nodeProxyAddresses, setNodeProxyAddresses] = useState<Map<string, string>>(new Map());
  const [nodeToSetup, setNodeToSetup] = useState<[string, string]>(null);
  // The status of each node. Key is the node's address.
  const [DKGStatuses, setDKGStatuses] = useState<Map<string, NodeStatus>>(new Map());

  const [nodeLoading, setNodeLoading] = useState<Map<string, boolean>>(null);
  const [DKGLoading, setDKGLoading] = useState(true);

  const ongoingItem = 'ongoingAction' + formID;
  const nodeToSetupItem = 'nodeToSetup' + formID;

  // called by a DKG row
  const notifyDKGState = (node: string, info: InternalDKGInfo) => {
    if (
      info.getStatus() === NodeStatus.Setup &&
      (status === Status.Initial || status === Status.Initialized)
    ) {
      setStatus(Status.Setup);
    }

    const newDKGStatuses = new Map(DKGStatuses);
    newDKGStatuses.set(node, info.getStatus());
    setDKGStatuses(newDKGStatuses);
  };

  // called by a DKG row
  const notifyLoading = (node: string, l: boolean) => {
    const newLoading = new Map(nodeLoading);
    newLoading.set(node, l);
    setNodeLoading(newLoading);
  };

  // Fetch result when available after a status change
  useEffect(() => {
    if (status === Status.ResultAvailable && isResultAvailable) {
      getResults(formID, setError, setResult, setIsResultSet).then();
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
      const node = Array.from(nodeProxyAddresses).find(([_node, proxy]) => proxy !== '');
      if (node !== undefined) {
        setNodeToSetup(Array.from(nodeProxyAddresses).find(([_node, proxy]) => proxy !== ''));
      }
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

  // Keep the "DKGLoading" state according to "nodeLoading". This state tells if
  // one of the element on the map is true.
  useEffect(() => {
    if (nodeLoading !== null) {
      const someNodeLoading = Array.from(nodeLoading.values()).includes(true);
      setDKGLoading(someNodeLoading);
      if (!someNodeLoading) {
        setOngoingAction(OngoingAction.None);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [nodeLoading]);

  // Update the status of the form if necessary
  useEffect(() => {
    if (status === Status.Initial) {
      if (DKGStatuses !== null && !DKGLoading) {
        const statuses = Array.from(DKGStatuses.values());

        // We want to update only if all nodes have already set their status
        if (statuses.length !== roster.length) {
          return;
        }

        // TODO: can be modified such that if the majority of the node are
        // initialized than the form status can still be set to initialized
        if (
          statuses.includes(NodeStatus.NotInitialized) ||
          statuses.includes(NodeStatus.Unreachable) ||
          statuses.includes(NodeStatus.Failed)
        ) {
          return;
        }

        setStatus(Status.Initialized);

        // Status Failed is handled by useChangeAction
      }
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [DKGStatuses, status, DKGLoading]);

  useEffect(() => {
    if (error !== null) {
      setTextModalError(t('errorRetrievingForm') + error.message);
      setShowModalError(true);
      setError(null);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [error]);
  const [titles, setTitles] = useState<any>({});
  useEffect(() => {
    try {
      if (configObj.Title === undefined) return;
      setTitles(configObj.Title);
    } catch (e) {
      setError(e.error);
    }
  }, [configObj]);

  return (
    <div className="w-[60rem] font-sans px-4 py-4">
      <Modal
        showModal={showModalError}
        setShowModal={setShowModalError}
        textModal={textModalError === null ? '' : textModalError}
        buttonRightText={t('close')}
        onClose={() => {
          window.location.href = ROUTE_FORM_INDEX;
        }}
      />
      {!loading ? (
        <>
          <div className="pt-8 text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
            {i18n.language === 'en' && titles.en}
            {i18n.language === 'fr' && titles.fr}
            {i18n.language === 'de' && titles.de}
          </div>

          <div className="pt-2 break-all">Form ID : {formId}</div>
          {status >= Status.Open &&
            status <= Status.Canceled &&
            voters !== null &&
            voters !== undefined && (
              <div className="break-all">{t('numVotes', { num: voters.length })}</div>
            )}
          <div className="py-6 pl-2">
            <div className="font-bold uppercase text-lg text-gray-700">{t('status')}</div>
            {DKGLoading && ongoingAction === OngoingAction.None && (
              <div className="px-2 pt-6">
                <LoadingButton>{t('statusLoading')}</LoadingButton>
              </div>
            )}
            {(!DKGLoading || ongoingAction !== OngoingAction.None) && (
              <div className="px-2 pt-6 flex justify-center">
                <StatusTimeline status={status} ongoingAction={ongoingAction} />
              </div>
            )}
          </div>
          <div className="py-4 pl-2 pb-8">
            <div className="font-bold uppercase text-lg text-gray-700 pb-2">{t('action')}</div>
            <div className="px-2">
              {DKGLoading && ongoingAction === OngoingAction.None && (
                <LoadingButton>{t('actionLoading')}</LoadingButton>
              )}{' '}
              {(!DKGLoading || ongoingAction !== OngoingAction.None) && (
                <Action
                  status={status}
                  formID={formID}
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
                />
              )}
            </div>
          </div>
          {voters !== null && voters !== undefined && voters.length > 0 && (
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
                formId={formId}
                nodeProxyAddresses={nodeProxyAddresses}
                setNodeProxyAddresses={setNodeProxyAddresses}
                ongoingAction={ongoingAction}
                notifyDKGState={notifyDKGState}
                nodeToSetup={nodeToSetup}
                notifyLoading={notifyLoading}
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

FormShow.propTypes = {
  location: PropTypes.any,
};

export default FormShow;
