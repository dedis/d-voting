import { Dialog, Transition } from '@headlessui/react';
import { CogIcon } from '@heroicons/react/outline';
import { FC, Fragment, useRef } from 'react';
import { useTranslation } from 'react-i18next';

type ChooseProxyModalProps = {
  roster: string[];
  showModal: boolean;
  nodeProxyAddresses: Map<string, string>;
  nodeToSetup: [string, string];
  setNodeToSetup: (node: [string, string]) => void;
  setShowModal: (show: boolean) => void;
  setUserConfirmedAction: (confirm: boolean) => void;
};

const ChooseProxyModal: FC<ChooseProxyModalProps> = ({
  roster,
  showModal,
  nodeProxyAddresses,
  nodeToSetup,
  setNodeToSetup,
  setShowModal,
  setUserConfirmedAction,
}) => {
  const { t } = useTranslation();
  const cancelButtonRef = useRef(null);

  const closeModal = () => {
    setShowModal(false);
  };

  const confirmChoice = () => {
    setUserConfirmedAction(true);
    closeModal();
  };

  const handleCheck = (node: string, proxy: string) => {
    if (proxy !== '') {
      setNodeToSetup([node, proxy]);
    }
  };

  const proxyCheckbox = () => {
    return (
      nodeToSetup !== null &&
      roster.map((node, index) => {
        const proxy = nodeProxyAddresses.get(node);
        const checkable = proxy !== '';

        return (
          <div className="flex items-center my-4 ml-4" key={node}>
            <input
              id={node + index}
              type="radio"
              className={`w-4 h-4 border-gray-300 cursor-pointer ${
                !checkable && 'cursor-not-allowed'
              }`}
              checked={node === nodeToSetup[0]}
              onChange={() => handleCheck(node, proxy)}
              disabled={!checkable}
            />
            <label
              htmlFor={node + index}
              className={`block ml-2 text-sm font-medium ${checkable && 'text-gray-700'} ${
                !checkable && 'text-gray-400'
              } cursor-pointer ${!checkable && 'cursor-not-allowed'}`}
              onChange={() => handleCheck(node, proxy)}>
              Node {index} ({node})
            </label>
          </div>
        );
      })
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
                  <div className="mx-auto flex-shrink-0 flex items-center justify-center h-12 w-12 rounded-full bg-[#ff0000] sm:mx-0 sm:h-10 sm:w-10">
                    <CogIcon className="h-6 w-6 text-[#ff0000]" aria-hidden="true" />
                  </div>
                  <div className="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left">
                    <Dialog.Title as="h3" className="text-lg leading-6 font-medium text-gray-900">
                      {t('nodeSetup')}
                    </Dialog.Title>
                    <div className="mt-2">
                      <p className="text-sm text-gray-500">{t('inputNodeSetup')}</p>
                    </div>
                    {proxyCheckbox()}
                  </div>
                </div>
              </div>
              <div className="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
                <button
                  type="button"
                  className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-[#ff0000] text-base font-medium text-white hover:bg-[#b51f1f] focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-[#ff0000] sm:ml-3 sm:w-auto sm:text-sm"
                  onClick={confirmChoice}>
                  {t('setupNode')}
                </button>
                <button
                  type="button"
                  className="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-[#ff0000] sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm"
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

export default ChooseProxyModal;
