import React, { FC, useContext, useEffect, useMemo, useState } from 'react';
import { ENDPOINT_USER_RIGHTS } from 'components/utils/Endpoints';
import { FlashContext, FlashLevel } from 'index';
import Loading from 'pages/Loading';
import { useTranslation } from 'react-i18next';
import { fetchCall } from '../../components/utils/fetchCall';
import AdminTable from './AdminTable';
import DKGTable from './DKGTable';
import * as endpoints from 'components/utils/Endpoints';

const Admin: FC = () => {
  const { t } = useTranslation();
  const fctx = useContext(FlashContext);
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [nodeProxyLoading, setNodeProxyLoading] = useState(true);
  const [nodeProxyObject, setNodeProxyObject] = useState({ Proxies: [] });
  const [nodeProxyError, setNodeProxyError] = useState(null);

  const abortController = useMemo(() => new AbortController(), []);

  useEffect(() => {
    fetchCall(
      endpoints.getProxiesAddresses,
      {
        method: 'GET',
        signal: abortController.signal,
      },
      setNodeProxyObject,
      setNodeProxyLoading
    ).catch((e) => {
      setNodeProxyError(e);
    });
  }, [abortController.signal]);

  const [nodeProxyAddresses, setNodeProxyAddresses] = useState<Map<string, string>>(null);

  useEffect(() => {
    if (nodeProxyError !== null) {
      fctx.addMessage(t('errorRetrievingProxy') + nodeProxyError.message, FlashLevel.Error);
      setNodeProxyError(null);
      setNodeProxyLoading(false);
    }

    if (nodeProxyObject !== null) {
      const newNodeProxyAddresses = new Map();

      const proxies = nodeProxyObject.Proxies;

      for (const [node, proxy] of Object.entries(proxies)) {
        newNodeProxyAddresses.set(node, proxy);
      }

      setNodeProxyAddresses(newNodeProxyAddresses);
      setNodeProxyLoading(false);
    }

    return () => {
      abortController.abort();
    };
  }, [abortController, fctx, t, nodeProxyObject, nodeProxyError]);

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
      <div className="mt-4 mb-8">
        <DKGTable
          nodeProxyAddresses={nodeProxyAddresses}
          setNodeProxyAddresses={setNodeProxyAddresses}
        />
      </div>
    </div>
  ) : (
    <Loading />
  );
};

export default Admin;
