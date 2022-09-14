import React, { FC, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import * as endpoints from 'components/utils/Endpoints';
import { DKGInfo, InternalDKGInfo, NodeProxyAddress, NodeStatus } from 'types/node';
import { ID } from 'types/configuration';
import DKGStatus from 'components/utils/DKGStatus';
import IndigoSpinnerIcon from './IndigoSpinnerIcon';
import { OngoingAction } from 'types/election';
import Modal from 'components/modal/Modal';
import { ExclamationCircleIcon } from '@heroicons/react/outline';
import { pollDKG } from './utils/PollStatus';

const POLLING_INTERVAL = 1000;
const MAX_ATTEMPTS = 10;

type DKGStatusRowProps = {
  electionId: ID;
  node: string;
  index: number;
  loading: Map<string, boolean>;
  setLoading: (loading: Map<string, boolean>) => void;
  nodeProxyAddresses: Map<string, string>;
  setNodeProxyAddresses: (nodeProxy: Map<string, string>) => void;
  DKGStatuses: Map<string, NodeStatus>;
  setDKGStatuses: (DKFStatuses: Map<string, NodeStatus>) => void;
  setTextModalError: (error: string) => void;
  setShowModalError: (show: boolean) => void;
  // notify to start initialization
  ongoingAction: OngoingAction;
  // notify the parent of the new state
  notifyDKGState: (node: string, info: InternalDKGInfo) => void;
  nodeToSetup: [string, string];
};

const DKGStatusRow: FC<DKGStatusRowProps> = ({
  electionId,
  node, // node is the node address, not the proxy
  index,
  loading,
  setLoading,
  nodeProxyAddresses,
  setNodeProxyAddresses,
  setTextModalError,
  setShowModalError,
  ongoingAction,
  notifyDKGState,
  nodeToSetup,
}) => {
  const { t } = useTranslation();
  const [proxy, setProxy] = useState(null);
  const [DKGLoading, setDKGLoading] = useState(true);
  const [status, setStatus] = useState<NodeStatus>(null);

  const [info, setInfo] = useState('');

  const abortController = new AbortController();
  const signal = abortController.signal;
  const TIMEOUT = 10000;
  const request = {
    method: 'GET',
    headers: {
      'Access-Control-Allow-Origin': '*',
    },
    signal: signal,
  };

  const [showModal, setShowModal] = useState(false);

  // Notify the parent each time our status changes
  useEffect(() => {
    notifyDKGState(node, InternalDKGInfo.fromStatus(status));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [status]);

  const initializeNode = async () => {
    const req = {
      method: 'POST',
      body: JSON.stringify({
        ElectionID: electionId,
        Proxy: proxy,
      }),
      headers: {
        'Content-Type': 'application/json',
      },
    };

    const response = await fetch(endpoints.dkgActors, req);
    if (!response.ok) {
      const txt = await response.text();
      throw new Error(txt);
    }
  };

  // Signal the start of DKG initialization
  useEffect(() => {
    if (
      ongoingAction === OngoingAction.Initializing &&
      (status === NodeStatus.NotInitialized || status === NodeStatus.Failed)
    ) {
      // TODO: can be modified such that if the majority of the node are
      // initialized than the election status can still be set to initialized

      setDKGLoading(true);

      initializeNode()
        .then(() => {
          setStatus(NodeStatus.Initialized);
        })
        .catch((e) => {
          setInfo(e.toString());
          setStatus(NodeStatus.Failed);
        });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ongoingAction]);

  const pollDKGStatus = (statusToMatch: NodeStatus) => {
    const req = {
      method: 'GET',
      signal: signal,
    };

    const match = (s: NodeStatus) => s === statusToMatch;

    return pollDKG(
      endpoints.getDKGActors(proxy, electionId),
      req,
      match,
      POLLING_INTERVAL,
      MAX_ATTEMPTS
    );
  };

  useEffect(() => {
    if (
      ongoingAction === OngoingAction.SettingUp &&
      nodeToSetup !== null &&
      nodeToSetup[0] === node
    ) {
      setDKGLoading(true);
      pollDKGStatus(NodeStatus.Setup)
        .then(
          () => {
            setStatus(NodeStatus.Setup);
          },
          (reason: any) => {
            setStatus(NodeStatus.Failed);
            setInfo(reason.toString());
          }
        )
        .catch((e) => {
          console.log('error:', e);
        });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ongoingAction]);

  // Set the mapping of the node and proxy address (only if the address was not
  // already fetched)
  useEffect(() => {
    if (node !== null && proxy === null) {
      const fetchNodeProxy = async () => {
        try {
          setTimeout(() => {
            abortController.abort();
          }, TIMEOUT);

          const response = await fetch(endpoints.getProxyAddress(node), request);

          if (!response.ok) {
            const js = await response.json();
            throw new Error(JSON.stringify(js));
          } else {
            let dataReceived = await response.json();
            return dataReceived as NodeProxyAddress;
          }
        } catch (e) {
          let errorMessage = t('errorRetrievingProxy');

          // Error triggered by timeout
          if (e instanceof DOMException) {
            errorMessage += t('proxyUnreachable', { node: node });
          } else {
            errorMessage += t('error');
          }

          setTextModalError(errorMessage + e.message);
          setShowModalError(true);

          // if we could not retrieve the proxy still resolve the promise
          // so that promise.then() goes to onSuccess() but display the error
          return { NodeAddr: node, Proxy: '' };
        }
      };

      setDKGLoading(true);

      fetchNodeProxy().then((nodeProxyAddress) => {
        setProxy(nodeProxyAddress.Proxy);
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [node, proxy]);

  useEffect(() => {
    if (proxy !== null) {
      // notify parent
      const newAddresses: Map<string, string> = new Map(nodeProxyAddresses);
      newAddresses.set(node, proxy);
      setNodeProxyAddresses(newAddresses);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [proxy]);

  // Fetch the status of the nodes
  useEffect(() => {
    if (proxy !== null && status === null) {
      const fetchDKGStatus = async (): Promise<InternalDKGInfo> => {
        // If we were not able to retrieve the proxy address of the node,
        // still return a resolved promise so that promise.then() goes to onSuccess().
        // Error was already displayed, no need to throw another one.
        if (proxy === '') {
          return InternalDKGInfo.fromStatus(NodeStatus.Unreachable);
        }

        try {
          setTimeout(() => {
            abortController.abort();
          }, TIMEOUT);

          const response = await fetch(endpoints.getDKGActors(proxy, electionId), request);

          if (response.status === 404) {
            return InternalDKGInfo.fromStatus(NodeStatus.NotInitialized);
          }

          if (!response.ok) {
            const js = await response.json();
            throw new Error(JSON.stringify(js));
          }

          let dataReceived = await response.json();
          console.log('data received:', dataReceived);
          return InternalDKGInfo.fromInfo(dataReceived as DKGInfo);
        } catch (e) {
          let errorMessage = t('errorRetrievingNodes');

          // Error triggered by timeout
          if (e instanceof DOMException) {
            errorMessage += t('nodeUnreachable', { node: node });
          } else {
            errorMessage += t('error');
          }

          setTextModalError(errorMessage + e.message);
          setShowModalError(true);

          // if we could not retrieve the proxy still resolve the promise
          // so that promise.then() goes to onSuccess() but display the error
          return InternalDKGInfo.fromStatus(NodeStatus.Unreachable);
        }
      };

      fetchDKGStatus().then((internalStatus) => {
        setStatus(internalStatus.getStatus());
        setInfo(internalStatus.getError());
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [proxy, status]);

  useEffect(() => {
    // UseEffect prevents the race condition on setDKGStatuses
    if (status !== null) {
      setDKGLoading(false);

      const newLoading = new Map(loading);
      newLoading.set(node, false);
      setLoading(newLoading);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [status]);

  return (
    <tr key={node} className="bg-white border-b hover:bg-gray-50">
      <td className="px-1.5 sm:px-6 py-4 font-medium text-gray-900 whitespace-nowrap truncate">
        {t('node')} {index} ({node})
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 flex flex-row">
        {!DKGLoading ? <DKGStatus status={status} /> : <IndigoSpinnerIcon />}
        <Modal
          textModal={info}
          buttonRightText="close"
          setShowModal={setShowModal}
          showModal={showModal}
        />
        {info !== '' && (
          <button
            onClick={() => {
              setShowModal(true);
            }}>
            <ExclamationCircleIcon className="ml-3 mr-2 h-5 w-5 stroke-orange-800" />
          </button>
        )}
      </td>
    </tr>
  );
};

export default DKGStatusRow;
