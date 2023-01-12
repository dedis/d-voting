import React, { FC, useContext, useEffect, useState } from 'react';
import { Dialog } from '@headlessui/react';
import SpinnerIcon from 'components/utils/SpinnerIcon';
import { CubeTransparentIcon } from '@heroicons/react/outline';
import { useTranslation } from 'react-i18next';
import usePostCall from 'components/utils/usePostCall';
import * as endpoints from 'components/utils/Endpoints';
import { FlashContext, FlashLevel } from 'index';
import AdminModal from './AdminModal';

type EditProxyModalProps = {
  open: boolean;
  setOpen: (open: boolean) => void;
  nodeProxy: Map<string, string>;
  setNodeProxy: (nodeProxy: Map<string, string>) => void;
  node: string;
  handleEditProxy(node: string, proxy: string): void;
};

const EditProxyModal: FC<EditProxyModalProps> = ({
  open,
  setOpen,
  nodeProxy,
  setNodeProxy,
  node,
  handleEditProxy,
}) => {
  const { t } = useTranslation();
  const fctx = useContext(FlashContext);
  const [currentProxy, setCurrentProxy] = useState('');
  const [currentNode, setCurrentNode] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [, setIsPosting] = useState(false);
  const [postError, setPostError] = useState(null);

  const sendFetchRequest = usePostCall(setPostError);

  useEffect(() => {
    if (postError !== null) {
      fctx.addMessage(t('editProxyError') + postError, FlashLevel.Error);
      setPostError(null);
    }
  }, [postError]);

  useEffect(() => {
    if (nodeProxy !== null) {
      setCurrentNode(node);
      setCurrentProxy(nodeProxy.get(node));
    }
  }, [nodeProxy, node]);

  const proxyAddressUpdate = async () => {
    const req = {
      method: 'PUT',
      body: JSON.stringify({
        Proxy: currentProxy,
        NewNode: currentNode,
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
  const handleNodeInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    setCurrentNode(e.target.value);
    setError(null);
  };

  const handleSave = async () => {
    setLoading(true);

    if (currentProxy !== '') {
      try {
        new URL(currentProxy);
        setError(null);
        const response = await proxyAddressUpdate();

        if (response) {
          const newNodeProxy = new Map(nodeProxy);
          newNodeProxy.set(currentNode, currentProxy);
          setNodeProxy(newNodeProxy);
          handleEditProxy(currentNode, currentProxy);
          fctx.addMessage(t('proxySuccessfullyEdited'), FlashLevel.Info);
        }

        setOpen(false);
      } catch {
        setError(t('invalidProxyError'));
      }
    } else {
      setError(t('inputProxyAddressError'));
    }
    setLoading(false);
  };

  const handleCancel = () => {
    setError(null);
    setOpen(false);
  };
  // TODO : modify other languages
  const modalBody = (
    <>
      <Dialog.Title as="h3" className="text-lg leading-6 font-medium text-gray-900">
        {t('editProxy')}
      </Dialog.Title>
      <div className="mt-10 mb-4 flex items-center">
        <label htmlFor={currentNode} className="mr-2">
          {t('node')}:{' '}
        </label>
        <input
          id={currentNode}
          type="text"
          className="border pl-2 w-1/2 py-1 flex rounded-lg"
          onChange={(e) => handleNodeInput(e)}
          placeholder="123.123..."
          value={currentNode}
        />
      </div>
      <div className="mt-10 mb-4 flex items-center">
        <label htmlFor={currentProxy} className="mr-2">
          {t('proxy')}:{' '}
        </label>
        <input
          id={currentProxy}
          type="text"
          className="border pl-2 w-1/2 py-1 flex rounded-lg"
          onChange={(e) => handleTextInput(e)}
          placeholder="https:// ..."
          value={currentProxy}
        />
        {error !== null && <div className="text-red-600 text-sm my-2">{error}</div>}
      </div>
    </>
  );

  const actionButton = (
    <button
      type="button"
      className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:col-start-2 sm:text-sm"
      onClick={() => handleSave()}>
      {loading ? (
        <SpinnerIcon />
      ) : (
        <CubeTransparentIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
      )}
      {t('save')}
    </button>
  );

  return (
    <AdminModal
      open={open}
      setOpen={setOpen}
      modalBody={modalBody}
      actionButton={actionButton}
      handleCancel={handleCancel}
    />
  );
};

export default EditProxyModal;
