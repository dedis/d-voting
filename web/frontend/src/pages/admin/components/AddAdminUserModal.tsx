import React, { FC, Fragment, useContext, useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { Dialog, Listbox, Transition } from '@headlessui/react';
import { CheckIcon, SelectorIcon } from '@heroicons/react/solid';

import { useTranslation } from 'react-i18next';
import SpinnerIcon from 'components/utils/SpinnerIcon';
import { UserAddIcon } from '@heroicons/react/outline';
import ShortUniqueId from 'short-unique-id';
import { FlashContext, FlashLevel } from 'index';
import { UserRole } from 'types/userRole';
import { ENDPOINT_ADD_ROLE } from 'components/utils/Endpoints';
import AdminModal from './AdminModal';
import usePostCall from 'components/utils/usePostCall';

const uid = new ShortUniqueId({ length: 8 });

type AddAdminUserModalProps = {
  open: boolean;
  setOpen(opened: boolean): void;
  handleAddRoleUser(user: object): void;
};

const roles = [UserRole.Admin, UserRole.Operator];

const AddAdminUserModal: FC<AddAdminUserModalProps> = ({ open, setOpen, handleAddRoleUser }) => {
  const { t } = useTranslation();
  const fctx = useContext(FlashContext);

  const [error, setError] = useState(null);
  const [postError, setPostError] = useState(null);
  const [, setIsPosting] = useState(false);
  const [loading, setLoading] = useState(false);
  const [sciperValue, setSciperValue] = useState('');
  const [selectedRole, setSelectedRole] = useState(roles[0]);

  const handleCancel = () => {
    setOpen(false)
    setError(null);
};
  const sendFetchRequest = usePostCall(setPostError);

  useEffect(() => {
    if (postError !== null) {
      fctx.addMessage(t('addRoleError') + postError, FlashLevel.Error);
      setPostError(null);
    }
  }, [postError]);
  const handleUserInput = (e: any) => {
    setSciperValue(e.target.value);
  };
  const userToAdd = { id: uid(), sciper: sciperValue, role: selectedRole };
  const saveMapping = async () => {
    const request = {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({userToAdd}),
    };
    return sendFetchRequest(ENDPOINT_ADD_ROLE, request, setIsPosting);
  };
  const handleAddUser = async () => {
    setLoading(true);
    if(sciperValue !== '' ) {
    try {
      setError(null);  
      const res = await saveMapping();
      console.log('res', res)
      if (!res) {
        setSciperValue('');
        setSelectedRole(roles[0]);
        handleAddRoleUser(userToAdd);
        console.log('sciperValue', sciperValue)
        fctx.addMessage(`${t('successAddUser')}`, FlashLevel.Info);
      }
      setOpen(false);
    } catch  {
        fctx.addMessage(`${t('errorAddRoleError')}`, FlashLevel.Error);
    }
    }else{
        setError(t('addRoleError'));
    }
    setLoading(false);
  };

  const modalBody = (
    <>
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
                          className={`block truncate ${selected ? 'font-medium' : 'font-normal'}`}>
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
    </>
  );

  const actionButton = (
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

AddAdminUserModal.propTypes = {
  open: PropTypes.bool.isRequired,
  setOpen: PropTypes.func.isRequired,
};

export default AddAdminUserModal;
