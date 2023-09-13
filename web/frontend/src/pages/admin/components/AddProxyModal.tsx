import React, { FC, useContext, useEffect, useState } from 'react';
import { Dialog } from '@headlessui/react';
import { PlusIcon } from '@heroicons/react/outline';
import SpinnerIcon from 'components/utils/SpinnerIcon';
import { FlashContext, FlashLevel } from 'index';
import { useTranslation } from 'react-i18next';
import * as endpoints from 'components/utils/Endpoints';
import usePostCall from 'components/utils/usePostCall';
import AdminModal from './AdminModal';

type AddProxyModalProps = {
  open: boolean;
  setOpen(opened: boolean): void;
  handleAddProxy(node: string, proxy: string): void;
};

const AddProxyModal: FC<AddProxyModalProps> = ({ open, setOpen, handleAddProxy }) => {
  const { t } = useTranslation();
  const fctx = useContext(FlashContext);
  const [error, setError] = useState(null);
  const [postError, setPostError] = useState(null);
  const [, setIsPosting] = useState(false);
  const [loading, setLoading] = useState(false);
  const [node, setNode] = useState('');
  const [proxy, setProxy] = useState('');

  const sendFetchRequest = usePostCall(setPostError);

  useEffect(() => {
    if (postError !== null) {
      fctx.addMessage(t('addNodeProxyError') + postError, FlashLevel.Error);
      setPostError(null);
    }
  }, [fctx, t, postError]);

  const handleNodeInput = (e: any) => {
    setNode(e.target.value);
    setError(null);
  };

  const handleProxyInput = (e: any) => {
    setProxy(e.target.value);
    setError(null);
  };

  const saveMapping = async () => {
    const request = {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        NodeAddr: node,
        Proxy: proxy,
      }),
    };
    return sendFetchRequest(endpoints.newProxyAddress, request, setIsPosting);
  };

  const handleAdd = async () => {
    setLoading(true);
    if (node !== '' && proxy !== '') {
      try {
        new URL(proxy);
        setError(null);

        const response = await saveMapping();

        if (response) {
          handleAddProxy(node, proxy);
          fctx.addMessage(t('nodeProxySuccessfullyAdded'), FlashLevel.Info);
          setNode('');
          setProxy('');
        }

        setOpen(false);
      } catch {
        setError(t('invalidProxyError'));
      }
    } else {
      setError(t('inputNodeProxyError'));
    }
    setLoading(false);
  };

  const handleCancel = () => {
    setError(null);
    setOpen(false);
  };

  const modalBody = (
    <>
      <Dialog.Title as="h3" className="text-lg leading-6 font-medium text-gray-900">
        {t('enterNodeProxy')}
      </Dialog.Title>
      <div className="mt-10 mb-4 flex items-center">
        <label htmlFor="node" className="mr-2">
          {t('node')}:{' '}
        </label>
        <input
          onChange={handleNodeInput}
          value={node}
          placeholder="123.123..."
          className="border pl-2 w-1/2 py-1 flex rounded-lg"
          id="node"
        />
      </div>

      <div className="mt-2 mb-10 flex items-center">
        <label htmlFor="proxy" className="mr-2">
          {t('proxy')}:{' '}
        </label>
        <input
          onChange={handleProxyInput}
          value={proxy}
          placeholder="https://..."
          className="border pl-2 w-1/2 py-1 flex rounded-lg"
          id="proxy"
        />
      </div>

      {error !== null && <div className="text-left text-red-600 text-sm my-2">{error}</div>}
    </>
  );

  const actionButton = (
    <>
      <button
        type="button"
        className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:col-start-2 sm:text-sm"
        onClick={handleAdd}>
        {loading ? <SpinnerIcon /> : <PlusIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />}
        {t('add')}
      </button>
    </>
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

export default AddProxyModal;
