import { MinusCircleIcon } from '@heroicons/react/outline';
import { Dialog, Transition } from '@headlessui/react';
import SpinnerIcon from 'components/utils/SpinnerIcon';
import { FlashContext, FlashLevel } from 'index';
import React, { FC, Fragment, useContext, useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import usePostCall from 'components/utils/usePostCall';
import * as endpoints from 'components/utils/Endpoints';

type RemoveProxyModalProps = {
  open: boolean;
  setOpen: (open: boolean) => void;
  node: string;
  handleDeleteProxy(): void;
};

const RemoveProxyModal: FC<RemoveProxyModalProps> = ({
  open,
  setOpen,
  node,
  handleDeleteProxy,
}) => {
  const { t } = useTranslation();
  const fctx = useContext(FlashContext);

  const [loading, setLoading] = useState(false);
  const [postError, setPostError] = useState(null);
  const [isPosting, setIsPosting] = useState(false);
  const cancelButtonRef = useRef(null);

  const sendFetchRequest = usePostCall(setPostError);

  useEffect(() => {
    if (postError !== null) {
      fctx.addMessage(t('removeProxyError') + postError, FlashLevel.Error);
      setPostError(null);
    }
  }, [postError]);

  const handleDelete = async () => {
    setLoading(true);

    const req = {
      method: 'DELETE',
    };

    const response = await sendFetchRequest(endpoints.editProxyAddress(node), req, setIsPosting);

    if (response) {
      handleDeleteProxy();
      fctx.addMessage(t('proxySuccessfullyDeleted'), FlashLevel.Info);
    }

    setOpen(false);
    setLoading(false);
  };

  return (
    <div>
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
              <div className="inline-block bg-white rounded-lg  text-left overflow-hidden shadow-xl transform transition-all my-8 align-middle max-w-lg w-full p-6">
                <div>
                  <div className="text-center">
                    <Dialog.Title as="h3" className="text-lg leading-6 font-medium text-gray-900">
                      {t('confirmDeleteProxy')}
                    </Dialog.Title>
                    <div className="my-4">
                      {t('node')}: {node}
                    </div>
                  </div>
                </div>
                <div className="mt-5 sm:mt-6 sm:grid sm:grid-cols-2 sm:gap-3 sm:grid-flow-row-dense">
                  <button
                    type="button"
                    className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:col-start-2 sm:text-sm"
                    onClick={handleDelete}>
                    {loading ? (
                      <SpinnerIcon />
                    ) : (
                      <MinusCircleIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
                    )}
                    {t('delete')}
                  </button>
                  <button
                    type="button"
                    className="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:col-start-1 sm:text-sm"
                    onClick={() => setOpen(false)}
                    ref={cancelButtonRef}>
                    {t('cancel')}
                  </button>
                </div>
              </div>
            </Transition.Child>
          </div>
        </Dialog>
      </Transition.Root>
    </div>
  );
};

export default RemoveProxyModal;
