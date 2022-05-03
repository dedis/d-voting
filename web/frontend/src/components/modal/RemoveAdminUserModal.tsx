import React, { FC, Fragment, useRef, useState } from 'react';
import { ENDPOINT_REMOVE_ROLE } from '../utils/Endpoints';
import PropTypes from 'prop-types';
import { Dialog, Transition } from '@headlessui/react';
import { UserRemoveIcon } from '@heroicons/react/outline';
import SpinnerIcon from 'components/utils/SpinnerIcon';
import { useTranslation } from 'react-i18next';

type RemoveAdminUserModalProps = {
  open: boolean;
  setOpen(opened: boolean): void;
  sciper: number;
};

const RemoveAdminUserModal: FC<RemoveAdminUserModalProps> = ({ open, setOpen, sciper }) => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);

  const handleClose = () => setOpen(false);

  const handleDelete = () => {
    const requestOptions = {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ sciper: sciper }),
    };
    setLoading(true);
    fetch(ENDPOINT_REMOVE_ROLE, requestOptions).then((data) => {
      setLoading(false);
      if (data.status === 200) {
        alert('User removed successfully');
        setOpen(false);
      } else {
        alert('Error while adding the user');
      }
    });
  };
  const cancelButtonRef = useRef(null);

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
                      {t('confirmDeleteUserSciper')} {sciper}
                    </Dialog.Title>
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
                      <UserRemoveIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
                    )}
                    {t('delete')}
                  </button>
                  <button
                    type="button"
                    className="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:col-start-1 sm:text-sm"
                    onClick={handleClose}
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

RemoveAdminUserModal.propTypes = {
  open: PropTypes.bool.isRequired,
  setOpen: PropTypes.func.isRequired,
  sciper: PropTypes.number.isRequired,
};

export default RemoveAdminUserModal;
