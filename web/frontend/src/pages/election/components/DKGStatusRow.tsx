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
const MAX_ATTEMPTS = 20;

type DKGStatusRowProps = {
  electionId: ID;
  node: string;
  index: number;
  nodeProxyAddresses: Map<string, string>;
  setNodeProxyAddresses: (nodeProxy: Map<string, string>) => void;
  setTextModalError: (error: string) => void;
  setShowModalError: (show: boolean) => void;
  // notify to start initialization
  ongoingAction: OngoingAction;
  // notify the parent of the new state
  notifyDKGState: (node: string, info: InternalDKGInfo) => void;
  // contains the node/proxy address of the node to setup
  nodeToSetup: [string, string];
  notifyLoading: (node: string, loading: boolean) => void;
};

const DKGStatusRow: FC<DKGStatusRowProps> = ({
  electionId,
  node, // node is the node address, not the proxy
  index,
  nodeProxyAddresses,
  setNodeProxyAddresses,
  setTextModalError,
  setShowModalError,
  ongoingAction,
  notifyDKGState,
  nodeToSetup,
  notifyLoading,
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

  // Notify the parent when we are loading or not
  useEffect(() => {
    notifyLoading(node, DKGLoading);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [DKGLoading]);

  // send the initialization request
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

  // Initialize the node if the initialization is ongoing and we are in a
  // legitimate status.
  useEffect(() => {
    if (
      ongoingAction === OngoingAction.Initializing &&
      (status === NodeStatus.NotInitialized || status === NodeStatus.Failed)
    ) {
      setDKGLoading(true);

      initializeNode()
        .then(() => {
          setStatus(NodeStatus.Initialized);
        })
        .catch((e: Error) => {
          setInfo(e.toString());
          setStatus(NodeStatus.Failed);
        })
        .finally(() => {
          setDKGLoading(false);
        });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ongoingAction]);

  const pollDKGStatus = (statusToMatch: NodeStatus): Promise<DKGInfo> => {
    const req = {
      method: 'GET',
      signal: signal,
    };

    const match = (s: NodeStatus) => s === statusToMatch;
    const statusUpdate = (s: NodeStatus) => setStatus(s);

    return pollDKG(
      endpoints.getDKGActors(proxy, electionId),
      req,
      match,
      POLLING_INTERVAL,
      MAX_ATTEMPTS,
      statusUpdate
    );
  };

  // Action taken when the setting up is triggered.
  useEffect(() => {
    if (ongoingAction !== OngoingAction.SettingUp || nodeToSetup === null) {
      return;
    }

    setDKGLoading(true);

    let expectedStatus = NodeStatus.Certified;
    if (nodeToSetup[0] === node) {
      expectedStatus = NodeStatus.Setup;
    }

    pollDKGStatus(expectedStatus)
      .then(
        () => {},
        (e: Error) => {
          setStatus(NodeStatus.Failed);
          setInfo(e.toString());
        }
      )
      .finally(() => setDKGLoading(false));
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
        setDKGLoading(false);
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [proxy, status]);

  return (
    <tr key={node} className="bg-white border-b hover:bg-gray-50">
      <td className="px-1.5 sm:px-6 py-4 font-medium text-gray-900 whitespace-nowrap truncate">
        {t('node')} {index} ({node})
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 flex flex-row">
        {DKGLoading && <IndigoSpinnerIcon />}
        <DKGStatus status={status} />
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
