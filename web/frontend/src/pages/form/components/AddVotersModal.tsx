import { Dialog, Transition } from '@headlessui/react';
import { CogIcon } from '@heroicons/react/outline';
import { FC, Fragment, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';

type AddVotersModalSuccessProps = {
  showModal: boolean;
  setShowModal: (show: boolean) => void;
  newVoters: string;
};

export const AddVotersModalSuccess: FC<AddVotersModalSuccessProps> = ({
  showModal,
  setShowModal,
  newVoters,
}) => {
  const { t } = useTranslation();

  function closeModal() {
    setShowModal(false);
  }

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
                      {t('addVotersDialog')}
                    </Dialog.Title>
                    <div className="mt-2">
                      <p className="text-sm text-gray-500">{t('votersAdded')}</p>
                    </div>
                    <pre>{newVoters}</pre>
                  </div>
                </div>
              </div>
              <div className="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
                <button
                  type="button"
                  className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-[#ff0000] text-base font-medium text-white hover:bg-[#b51f1f] focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-[#ff0000] sm:ml-3 sm:w-auto sm:text-sm"
                  onClick={closeModal}>
                  {t('confirm')}
                </button>
              </div>
            </div>
          </Transition.Child>
        </div>
      </Dialog>
    </Transition.Root>
  );
};
type AddVotersModalProps = {
  showModal: boolean;
  setShowModal: (show: boolean) => void;
  setUserConfirmedAction: (voters: string) => void;
};

export const AddVotersModal: FC<AddVotersModalProps> = ({
  showModal,
  setShowModal,
  setUserConfirmedAction,
}) => {
  const { t } = useTranslation();
  const cancelButtonRef = useRef(null);
  const [voters, setVoters] = useState('');

  const cancelModal = () => {
    setUserConfirmedAction('');
    setShowModal(false);
  };

  const confirmChoice = () => {
    setUserConfirmedAction(voters);
    setShowModal(false);
  };

  const votersBox = () => {
    return (
      <div>
        <textarea
          autoFocus={true}
          onChange={(e) => setVoters(e.target.value)}
          name="Voters"
          placeholder="SCIPERs"
          className="m-3 px-1 w-100 text-lg border rounded-md"
          rows={10}
        />
      </div>
    );
  };

  return (
    <Transition.Root show={showModal} as={Fragment}>
      <Dialog as="div" className="fixed z-10 inset-0 px-4 overflow-y-auto" onClose={cancelModal}>
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
                      {t('addVotersDialog')}
                    </Dialog.Title>
                    <div className="mt-2">
                      <p className="text-sm text-gray-500">{t('inputAddVoters')}</p>
                    </div>
                    {votersBox()}
                  </div>
                </div>
              </div>
              <div className="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
                <button
                  data-testid="addVotersConfirm"
                  type="button"
                  className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-[#ff0000] text-base font-medium text-white hover:bg-[#b51f1f] focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-[#ff0000] sm:ml-3 sm:w-auto sm:text-sm"
                  onClick={confirmChoice}>
                  {t('addVotersConfirm')}
                </button>
                <button
                  type="button"
                  className="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-[#ff0000] sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm"
                  onClick={cancelModal}
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
