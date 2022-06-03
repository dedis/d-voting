import React, { FC, Fragment, useContext, useEffect, useRef, useState } from 'react';
import { Dialog, Transition } from '@headlessui/react';
import { PlusIcon } from '@heroicons/react/outline';
import SpinnerIcon from 'components/utils/SpinnerIcon';
import { FlashContext, FlashLevel } from 'index';
import { useTranslation } from 'react-i18next';
import * as endpoints from 'components/utils/Endpoints';
import usePostCall from 'components/utils/usePostCall';

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
  const cancelButtonRef = useRef(null);

  const sendFetchRequest = usePostCall(setPostError);

  useEffect(() => {
    if (postError !== null) {
      fctx.addMessage(t('addNodeProxyError') + postError, FlashLevel.Error);
      setPostError(null);
    }
  }, [postError]);

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

  return (
    <Transition.Root show={open} as={Fragment}>
      <Dialog
        as="div"
        className="fixed z-10 inset-0 px-4 sm:px-0 overflow-y-auto"
        initialFocus={cancelButtonRef}
        onClose={setOpen}>
        <div className="block items-end justify-center min-h-screen text-center">
          <Transition.Child
            as={Fragment}
            enter="ease-out duration-300"
            enterFrom="opacity-0"
            enterTo="opacity-100"
            leave="ease-in duration-200"
            leaveFrom="opacity-100"
            leaveTo="opacity-0">
            <Dialog.Overlay className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
          </Transition.Child>

          {/* This element is to trick the browser into centering the modal contents. */}
          <span className="inline-block align-middle h-screen" aria-hidden="true">
            &#8203;
          </span>
          <Transition.Child
            as={Fragment}
            enter="ease-out duration-300"
            enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
            enterTo="opacity-100 translate-y-0 sm:scale-100"
            leave="ease-in duration-200"
            leaveFrom="opacity-100 translate-y-0 sm:scale-100"
            leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95">
            <div className="inline-block bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all my-8 align-middle max-w-lg w-full p-6">
              <div>
                <div className="text-center">
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

                  {error !== null && (
                    <div className="text-left text-red-600 text-sm my-2">{error}</div>
                  )}
                </div>
              </div>
              <div className="mt-5 sm:mt-6 sm:grid sm:grid-cols-2 sm:gap-3 sm:grid-flow-row-dense">
                <button
                  type="button"
                  className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:col-start-2 sm:text-sm"
                  onClick={handleAdd}>
                  {loading ? (
                    <SpinnerIcon />
                  ) : (
                    <PlusIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
                  )}
                  {t('add')}
                </button>
                <button
                  type="button"
                  className="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:col-start-1 sm:text-sm"
                  onClick={handleCancel}
                  ref={cancelButtonRef}>
                  {t('cancel')}
                </button>
              </div>
            </div>
          </Transition.Child>
        </div>
      </Dialog>
    </Transition.Root>
  );
};

export default AddProxyModal;
