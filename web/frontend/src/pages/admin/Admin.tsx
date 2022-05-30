import React, { FC, useContext, useEffect, useState } from 'react';
import { ENDPOINT_USER_RIGHTS } from 'components/utils/Endpoints';
import { FlashContext, FlashLevel } from 'index';
import Loading from 'pages/Loading';
import { useTranslation } from 'react-i18next';
import AdminTable from './AdminTable';
import DKGTable from './DKGTable';
import useFetchCall from 'components/utils/useFetchCall';
import * as endpoints from 'components/utils/Endpoints';

const Admin: FC = () => {
  const { t } = useTranslation();
  const fctx = useContext(FlashContext);
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);

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
  const [, setNodeProxyLoading] = useState(true);

  const [nodeProxyAddresses, setNodeProxyAddresses] = useState<Map<string, string>>(null);

  useEffect(() => {
    if (nodeProxyError !== null) {
      fctx.addMessage(t('errorRetrievingProxy') + nodeProxyError.message, FlashLevel.Error);
      setNodeProxyLoading(false);
    }

    if (nodeProxyObject !== null) {
      const newNodeProxyAddresses = new Map();

      nodeProxyObject.Proxies.forEach((value) => {
        Object.entries(value).forEach(([node, proxy]) => {
          newNodeProxyAddresses.set(node, proxy);
        });
      });

      setNodeProxyAddresses(newNodeProxyAddresses);
      setNodeProxyLoading(false);
    }

    return () => {
      abortController.abort();
    };
  }, [nodeProxyObject, nodeProxyError]);

  useEffect(() => {
    fetch(ENDPOINT_USER_RIGHTS)
      .then((resp) => {
        setLoading(false);
        if (resp.status === 200) {
          const jsonData = resp.json();
          jsonData.then((result) => {
            setUsers(result);
          });
        } else {
          setUsers([]);
          fctx.addMessage(t('errorFetchingUsers'), FlashLevel.Error);
        }
      })
      .catch((error) => {
        setLoading(false);
        fctx.addMessage(`${t('errorFetchingUsers')}: ${error.message}`, FlashLevel.Error);
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return !loading && !nodeProxyLoading ? (
    <div className="w-[60rem] font-sans px-4 py-4">
      <div className="flex items-center justify-between mb-4 pt-8">
        <div className="flex-1 min-w-0">
          <h2 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
            {t('admin')}
          </h2>
        </div>
      </div>

      <AdminTable users={users} setUsers={setUsers} />

      <div className="py-6 pl-2">
        <div className="font-bold uppercase text-lg text-gray-700">{t('DKGStatuses')}</div>
      </div>

      <DKGTable nodeProxyAddresses={nodeProxyAddresses} />
    </div>
  ) : (
    <Loading />
  );
};

export default Admin;
