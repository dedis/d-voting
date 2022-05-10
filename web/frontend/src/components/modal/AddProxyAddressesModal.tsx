import { Dialog, Transition } from '@headlessui/react';
import { CubeTransparentIcon } from '@heroicons/react/outline';
import { FC, Fragment, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';

type AddProxyAddressesModalProps = {
  roster: string[];
  showModal: boolean;
  proxyAddresses: Map<string, string>;
  setProxyAddresses: (addresses: Map<string, string>) => void;
  setShowModal: (show: boolean) => void;
  setUserConfirmedAction: (confirm: boolean) => void;
};

const AddProxyAddressesModal: FC<AddProxyAddressesModalProps> = ({
  roster,
  showModal,
  proxyAddresses,
  setProxyAddresses,
  setShowModal,
  setUserConfirmedAction,
}) => {
  const { t } = useTranslation();
  const cancelButtonRef = useRef(null);
  const [error, setError] = useState(null);

  const closeModal = () => {
    setError(null);
    setShowModal(false);
  };

  const confirmChoice = () => {
    setError(null);
    if (proxyAddresses && !Array.from(proxyAddresses.values()).includes('')) {
      setProxyAddresses(proxyAddresses);
      setUserConfirmedAction(true);
      closeModal();
    } else {
      setError(t('inputProxyAddressesError'));
    }
  };

  const handleTextInput = (e: React.ChangeEvent<HTMLInputElement>, node: string) => {
    const newAddresses = new Map(proxyAddresses);
    newAddresses.set(node, e.target.value);

    setProxyAddresses(newAddresses);
  };

  const proxyInputField = () => {
    return (
      <div>
        {roster.map((node) => (
          <div key={node} className="flex justify-between py-2">
            <label htmlFor={node}>Proxy: </label>
            <input
              id={node}
              type="text"
              className="flex-auto ml-4 sm:text-md border rounded-md text-gray-600"
              onChange={(e) => handleTextInput(e, node)}
              placeholder={proxyAddresses.get(node) === '' ? 'https:// ...' : ''}
              value={proxyAddresses.get(node)}
            />
          </div>
        ))}
        <div className="text-red-600 text-sm py-2 sm:pl-2 pl-1">{error}</div>
      </div>
    );
  };

  return (
    <Transition.Root show={showModal} as={Fragment}>
      <Dialog as="div" className="fixed z-10 inset-0 px-4 overflow-y-auto" onClose={closeModal}>
        <div className="block items-end justify-center min-h-screen text-center">
          <Dialog.Overlay className="fixed inset-0 bg-black opacity-30" />

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
            <div className=" inline-block bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all my-8 align-middle max-w-lg w-full">
              <div className="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
                <div className="sm:flex sm:items-start">
                  <div className="mx-auto flex-shrink-0 flex items-center justify-center h-12 w-12 rounded-full bg-indigo-100 sm:mx-0 sm:h-10 sm:w-10">
                    <CubeTransparentIcon className="h-6 w-6 text-indigo-600" aria-hidden="true" />
                  </div>
                  <div className="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left">
                    <Dialog.Title as="h3" className="text-lg leading-6 font-medium text-gray-900">
                      {t('nodeInitialization')}
                    </Dialog.Title>
                    <div className="mt-2">
                      <p className="text-sm text-gray-500">{t('inputProxyAddresses')}</p>
                    </div>
                    {proxyInputField()}
                  </div>
                </div>
              </div>
              <div className="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
                <button
                  type="button"
                  className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:ml-3 sm:w-auto sm:text-sm"
                  onClick={confirmChoice}>
                  {t('initializeNode')}
                </button>
                <button
                  type="button"
                  className="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm"
                  onClick={closeModal}
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

export default AddProxyAddressesModal;
