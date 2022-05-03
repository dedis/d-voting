import React, { FC, Fragment, useRef, useState } from 'react';
import PropTypes from 'prop-types';
import { Dialog, Listbox, Transition } from '@headlessui/react';
import { CheckIcon, SelectorIcon } from '@heroicons/react/solid';

import { ENDPOINT_ADD_ROLE } from 'components/utils/Endpoints';
import { useTranslation } from 'react-i18next';
import SpinnerIcon from 'components/utils/SpinnerIcon';
import { UserAddIcon } from '@heroicons/react/outline';

type AddAdminUserModalProps = {
  open: boolean;
  setOpen(opened: boolean): void;
};

const roles = ['Admin', 'Operator'];

const AddAdminUserModal: FC<AddAdminUserModalProps> = ({ open, setOpen }) => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [sciperValue, setSciperValue] = useState('');
  const [selectedRole, setSelectedRole] = useState(roles[0]);

  const handleClose = () => setOpen(false);

  const handleUserInput = (e: any) => {
    setSciperValue(e.target.value);
  };

  const handleAddUser = () => {
    const requestOptions = {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ sciper: sciperValue, role: selectedRole }),
    };
    setLoading(true);
    fetch(ENDPOINT_ADD_ROLE, requestOptions).then((data) => {
      setLoading(false);
      if (data.status === 200) {
        alert('User added successfully');
        setOpen(false);
      } else {
        alert('Error while adding the user');
      }
    });
  };
  const cancelButtonRef = useRef(null);

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
                    {t('enterSciper')}
                  </Dialog.Title>
                  <input
                    onChange={handleUserInput}
                    value={sciperValue}
                    placeholder="Sciper"
                    className="mt-8 mb-4 border pl-2 w-1/2 py-1 flex rounded-lg"
                  />
                  <div className="mt-2 pb-4">
                    <Listbox value={selectedRole} onChange={setSelectedRole}>
                      <div className="relative mt-1">
                        <Listbox.Button className="relative w-full cursor-default rounded-lg bg-white py-2 pl-3 pr-10 text-left border focus:outline-none focus-visible:border-indigo-500 focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-orange-300 sm:text-sm">
                          <span className="block truncate">{selectedRole}</span>
                          <span className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2">
                            <SelectorIcon className="h-5 w-5 text-gray-400" aria-hidden="true" />
                          </span>
                        </Listbox.Button>
                        <Transition
                          as={Fragment}
                          leave="transition ease-in duration-100"
                          leaveFrom="opacity-100"
                          leaveTo="opacity-0">
                          <Listbox.Options className="absolute mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                            {roles.map((role, personIdx) => (
                              <Listbox.Option
                                key={personIdx}
                                className={({ active }) =>
                                  `relative cursor-default select-none py-2 pl-10 pr-4 ${
                                    active ? 'bg-indigo-100 text-indigo-900' : 'text-gray-900'
                                  }`
                                }
                                value={role}>
                                {({ selected }) => (
                                  <>
                                    <span
                                      className={`block truncate ${
                                        selected ? 'font-medium' : 'font-normal'
                                      }`}>
                                      {role}
                                    </span>
                                    {selected ? (
                                      <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-indigo-600">
                                        <CheckIcon className="h-5 w-5" aria-hidden="true" />
                                      </span>
                                    ) : null}
                                  </>
                                )}
                              </Listbox.Option>
                            ))}
                          </Listbox.Options>
                        </Transition>
                      </div>
                    </Listbox>
                  </div>
                </div>
              </div>
              <div className="mt-5 sm:mt-6 sm:grid sm:grid-cols-2 sm:gap-3 sm:grid-flow-row-dense">
                <button
                  type="button"
                  className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:col-start-2 sm:text-sm"
                  onClick={handleAddUser}>
                  {loading ? (
                    <SpinnerIcon />
                  ) : (
                    <UserAddIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
                  )}
                  {t('addUser')}
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
  );
};

AddAdminUserModal.propTypes = {
  open: PropTypes.bool.isRequired,
  setOpen: PropTypes.func.isRequired,
};

export default AddAdminUserModal;
