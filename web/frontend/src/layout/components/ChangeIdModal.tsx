import React, { FC, useContext, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Dialog } from '@headlessui/react';
import { useTranslation } from 'react-i18next';
import { ENDPOINT_LOGOUT } from '../../components/utils/Endpoints';
import { AuthContext, FlashContext, FlashLevel } from '../../index';
import handleLogin from '../../pages/session/HandleLogin';

type ChangeIdModalProps = {
  isShown: boolean;
  setIsShown: (isShown: boolean) => void;
};

const logout = async (t, authCtx, fctx, navigate) => {
  const opts = { method: 'POST' };

  const res = await fetch(ENDPOINT_LOGOUT, opts);
  if (res.status !== 200) {
    fctx.addMessage(t('logOutError', { error: res.statusText }), FlashLevel.Error);
  } else {
    fctx.addMessage(t('logOutSuccessful'), FlashLevel.Info);
  }
  // TODO: should be a setAuth function passed to AuthContext rather than
  // changing the state directly
  authCtx.isLogged = false;
  authCtx.firstName = undefined;
  authCtx.lastName = undefined;
  navigate('/');
};

const ChangeIdModal: FC<ChangeIdModalProps> = ({ isShown, setIsShown }) => {
  const { t } = useTranslation();
  const [newId, setNewId] = useState('');
  const authCtx = useContext(AuthContext);

  const navigate = useNavigate();

  const fctx = useContext(FlashContext);

  async function changeId() {
    logout(t, authCtx, fctx, navigate).then(() => {
      handleLogin(fctx, newId).catch(console.error);
    });
  }

  if (isShown) {
    return (
      <Dialog open={isShown} onClose={() => {}}>
        <Dialog.Overlay className="fixed inset-0 bg-black opacity-30" />
        <div className="fixed inset-0 flex items-center justify-center">
          <div className="bg-white content-center rounded-lg shadow-lg p-3 w-80">
            <Dialog.Title as="h3" className="text-lg font-medium leading-6 text-gray-900">
              {t('changeIdTitle')}
            </Dialog.Title>
            <Dialog.Description className="mt-2 mx-auto text-sm text-gray-500">
              {t('changeIdDialog')}
            </Dialog.Description>
            <div className="mt-4 sm:mt-6 sm:grid sm:grid-cols-2 sm:gap-3 sm:grid-flow-row-dense">
              <input
                onChange={(e) => setNewId(e.target.value)}
                type="number"
                placeholder={t('changeIdPlaceholder')}
                autoFocus={true}
              />
            </div>
            <div className="mt-4 sm:mt-6 sm:grid sm:grid-cols-2 sm:gap-3 sm:grid-flow-row-dense">
              <button
                type="button"
                className="inline-flex justify-center px-4 py-2 text-sm font-medium text-white bg-[#ff0000] border border-transparent rounded-md hover:bg-gray-300 focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-indigo-500"
                onClick={() => setIsShown(false)}>
                {t('cancel')}
              </button>
              <button
                type="button"
                className="inline-flex justify-center px-4 py-2 text-sm font-medium text-white bg-red-600 border border-transparent rounded-md hover:bg-red-700 focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-red-500"
                onClick={() => {
                  setIsShown(false);
                  changeId();
                }}>
                {t('changeIdContinue')}
              </button>
            </div>
          </div>
        </div>
      </Dialog>
    );
  } else {
    return <></>;
  }
};

export default ChangeIdModal;
