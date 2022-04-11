import { FC, Fragment } from 'react';
import { useNavigate } from 'react-router-dom';
import { Dialog, Transition } from '@headlessui/react';

type RedirectToModalProps = {
  showModal: boolean;
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>;
  title: string;
  children: string;
  buttonRightText: string;
  navigateDestination?: string;
};

const RedirectToModal: FC<RedirectToModalProps> = ({
  showModal,
  setShowModal,
  title,
  children,
  buttonRightText,
  navigateDestination,
}) => {
  const navigate = useNavigate();

  const closeModal = () => {
    if (navigateDestination) {
      navigate(navigateDestination);
    }
    setShowModal(false);
  };

  return (
    <Transition appear show={showModal} as={Fragment}>
      <Dialog as="div" className="fixed inset-0 z-10 overflow-y-auto" onClose={closeModal}>
        <div className="min-h-screen px-4 text-center">
          <Dialog.Overlay className="fixed inset-0 bg-black opacity-30" />

          <Transition.Child
            as={Fragment}
            enter="ease-out duration-300"
            enterFrom="opacity-0"
            enterTo="opacity-100"
            leave="ease-in duration-200"
            leaveFrom="opacity-100"
            leaveTo="opacity-0">
            <Dialog.Overlay className="fixed inset-0" />
          </Transition.Child>

          <span className="inline-block h-screen align-middle" aria-hidden="true">
            &#8203;
          </span>
          <Transition.Child
            as={Fragment}
            enter="ease-out duration-300"
            enterFrom="opacity-0 scale-95"
            enterTo="opacity-100 scale-100"
            leave="ease-in duration-200"
            leaveFrom="opacity-100 scale-100"
            leaveTo="opacity-0 scale-95">
            <div className="inline-block w-full max-w-md p-6 my-8 overflow-hidden text-left align-middle transition-all transform bg-white shadow-xl rounded-2xl">
              <Dialog.Title as="h3" className="text-lg font-medium leading-6 text-gray-900">
                {title}
              </Dialog.Title>
              <div className="mt-2">
                <p className="text-sm text-gray-500">{children}</p>
              </div>

              <div className="mt-4">
                <button
                  type="button"
                  className="inline-flex justify-center px-4 py-2 text-sm font-medium text-white bg-indigo-500 border border-transparent rounded-md hover:bg-indigo-400 focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-indigo-500"
                  onClick={closeModal}>
                  {buttonRightText}
                </button>
              </div>
            </div>
          </Transition.Child>
        </div>
      </Dialog>
    </Transition>
  );
};

export default RedirectToModal;
