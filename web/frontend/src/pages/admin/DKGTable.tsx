import React, { FC, useContext, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import ProxyRow from './ProxyRow';
import * as endpoints from 'components/utils/Endpoints';
import useFetchCall from 'components/utils/useFetchCall';
import { FlashContext, FlashLevel } from 'index';
import Loading from 'pages/Loading';

const DKGTable: FC = () => {
  const { t } = useTranslation();
  const fcxt = useContext(FlashContext);

  const abortController = new AbortController();
  const signal = abortController.signal;

  const request = {
    method: 'GET',
    signal: signal,
  };

  const [nodeProxyObject, nodeProxyLoading, nodeProxyError] = useFetchCall(
    endpoints.getProxiesAddresses,
    request
  );

  const [nodeProxyAddresses, setNodeProxyAddresses] = useState(null);

  useEffect(() => {
    if (nodeProxyError !== null) {
      fcxt.addMessage(nodeProxyError.message, FlashLevel.Error);
    }

    if (nodeProxyObject !== null) {
      const newNodeProxyAddresses = new Map();

      nodeProxyObject.Proxies.forEach((value) => {
        Object.entries(value).forEach(([node, proxy]) => {
          newNodeProxyAddresses.set(node, proxy);
        });
      });

      setNodeProxyAddresses(newNodeProxyAddresses);
    }

    return () => {
      abortController.abort();
    };
  }, [nodeProxyObject, nodeProxyError]);

  return !nodeProxyLoading ? (
    <div className="relative divide-y overflow-x-auto shadow-md sm:rounded-lg mt-2">
      <table className="w-full text-sm text-left text-gray-500">
        <thead className="text-xs text-gray-700 uppercase bg-gray-50">
          <tr>
            <th scope="col" className="px-6 py-3">
              {t('nodes')}
            </th>
            <th scope="col" className="px-6 py-3">
              {t('proxies')}
            </th>
            <th scope="col" className="px-6 py-3">
              <span className="sr-only">{t('edit')}</span>
            </th>
          </tr>
        </thead>
        <tbody>
          <>
            {nodeProxyAddresses !== null &&
              Array.from(nodeProxyAddresses).map(([node, proxy], index) => (
                <ProxyRow node={node} proxy={proxy} index={index} />
              ))}
          </>
        </tbody>
      </table>
    </div>
  ) : (
    <Loading />
  );
};

export default DKGTable;
