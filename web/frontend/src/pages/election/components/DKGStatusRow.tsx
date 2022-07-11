import React, { FC, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import * as endpoints from 'components/utils/Endpoints';
import { NodeProxyAddress, NodeStatus } from 'types/node';
import { ID } from 'types/configuration';
import DKGStatus from 'components/utils/DKGStatus';
import IndigoSpinnerIcon from './IndigoSpinnerIcon';

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
};

const DKGStatusRow: FC<DKGStatusRowProps> = ({
  electionId,
  node,
  index,
  loading,
  setLoading,
  nodeProxyAddresses,
  setNodeProxyAddresses,
  DKGStatuses,
  setDKGStatuses,
  setTextModalError,
  setShowModalError,
}) => {
  const { t } = useTranslation();
  const [proxy, setProxy] = useState(null);
  const [DKGLoading, setDKGLoading] = useState(true);
  const [status, setStatus] = useState<NodeStatus>(null);

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

  useEffect(() => {
    // update status on useChangeAction (i.e. initializing and setting up)
    if (DKGStatuses.has(node)) {
      setStatus(DKGStatuses.get(node));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [DKGStatuses]);

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
      const fetchDKGStatus = async () => {
        // If we were not able to retrieve the proxy address of the node,
        // still return a resolved promise so that promise.then() goes to onSuccess().
        // Error was already displayed, no need to throw another one.
        if (proxy === '') {
          return NodeStatus.Unreachable;
        }

        try {
          setTimeout(() => {
            abortController.abort();
          }, TIMEOUT);

          const response = await fetch(endpoints.getDKGActors(proxy, electionId), request);

          if (response.status === 404) {
            return NodeStatus.NotInitialized;
          }

          if (!response.ok) {
            const js = await response.json();
            throw new Error(JSON.stringify(js));
          }

          let dataReceived = await response.json();
          return dataReceived.Status as NodeStatus;
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
          return NodeStatus.Unreachable;
        }
      };

      fetchDKGStatus().then((nodeStatus) => {
        setStatus(nodeStatus);
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [proxy, status]);

  useEffect(() => {
    // UseEffect prevents the race condition on setDKGStatuses
    if (status !== null) {
      setDKGLoading(false);

      // notify parent
      const newDKGStatuses = new Map(DKGStatuses);
      newDKGStatuses.set(node, status);
      setDKGStatuses(newDKGStatuses);

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
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
        {!DKGLoading ? <DKGStatus status={status} /> : <IndigoSpinnerIcon />}
      </td>
    </tr>
  );
};

export default DKGStatusRow;
