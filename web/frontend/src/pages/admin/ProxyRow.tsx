import { AuthContext, FlashContext, FlashLevel } from 'index';
import React, { FC, useContext, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import * as endpoints from 'components/utils/Endpoints';
import { UserRole } from 'types/userRole';
import usePostCall from 'components/utils/usePostCall';

type ProxyRowProps = {
  node: string;
  proxy: string;
  index: number;
};

const ProxyRow: FC<ProxyRowProps> = ({ node, proxy, index }) => {
  const { t } = useTranslation();
  const fctx = useContext(FlashContext);
  const authCtx = useContext(AuthContext);

  const isAuthorized =
    authCtx.isLogged && (authCtx.role === UserRole.Admin || authCtx.role === UserRole.Operator);

  const [currentProxy, setCurrentProxy] = useState(null);
  const [previousProxy, setPreviousProxy] = useState(null);
  const [isEditMode, setIsEditMode] = useState(false);
  const [error, setError] = useState(null);
  const [isPosting, setIsPosting] = useState(false);
  const [postError, setPostError] = useState(null);

  const sendFetchRequest = usePostCall(setPostError);

  useEffect(() => {
    if (postError !== null) {
      fctx.addMessage(postError, FlashLevel.Error);
      setPostError(null);
    }
  }, [postError]);

  useEffect(() => {
    if (proxy !== null) {
      setCurrentProxy(proxy);
      setPreviousProxy(proxy);
    }
  }, [proxy]);

  const proxyAddressUpdate = async () => {
    const req = {
      method: 'PUT',
      body: JSON.stringify({
        Proxy: currentProxy,
      }),
      headers: {
        'Content-Type': 'application/json',
      },
    };
    return sendFetchRequest(endpoints.editProxyAddress(node), req, setIsPosting);
  };

  const handleTextInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    setCurrentProxy(e.target.value);
    setError(null);
  };

  const handleEdit = () => {
    setIsEditMode(true);
  };

  const handleSave = () => {
    if (proxy !== '') {
      setIsEditMode(false);
      setError(null);
      setPreviousProxy(currentProxy);
      proxyAddressUpdate();
    } else {
      setError(t('inputProxyAddressError'));
    }
  };

  const handleCancel = () => {
    setCurrentProxy(previousProxy);
    setIsEditMode(false);
    setError(null);
  };

  return (
    <tr key={node} className="bg-white border-b">
      <td scope="row" className="px-6 py-4 font-medium text-gray-600 whitespace-nowrap">
        {t('node')} {index} ({node})
      </td>
      <td className="px-6 py-4">
        {isEditMode ? (
          <>
            <input
              type="text"
              className="flex-auto sm:text-md border rounded-md text-gray-600"
              onChange={(e) => handleTextInput(e)}
              placeholder={currentProxy === '' ? 'https:// ...' : ''}
              value={currentProxy}
            />
            <>{error !== null && <div className="text-red-600 text-sm py-2">{error}</div>}</>
          </>
        ) : (
          <>{currentProxy}</>
        )}
      </td>
      <td className="px-6 py-4 text-right">
        {isAuthorized && (
          <>
            {isEditMode ? (
              <div>
                <button
                  onClick={() => handleSave()}
                  className="font-medium text-indigo-600 hover:underline mx-2">
                  {t('save')}
                </button>
                <button
                  onClick={() => handleCancel()}
                  className="font-medium text-indigo-600 hover:underline">
                  {t('cancel')}
                </button>
              </div>
            ) : (
              <button
                onClick={() => handleEdit()}
                className="font-medium text-indigo-600 hover:underline">
                {t('edit')}
              </button>
            )}
          </>
        )}
      </td>
    </tr>
  );
};

export default ProxyRow;
