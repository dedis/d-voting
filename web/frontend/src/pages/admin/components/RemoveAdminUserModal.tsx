import React, { FC, useContext, useRef, useState } from 'react';
import { ENDPOINT_REMOVE_ROLE } from 'components/utils/Endpoints';
import PropTypes from 'prop-types';
import { Dialog } from '@headlessui/react';
import { UserRemoveIcon } from '@heroicons/react/outline';
import SpinnerIcon from 'components/utils/SpinnerIcon';
import { useTranslation } from 'react-i18next';
import { FlashContext, FlashLevel } from 'index';
import AdminModal from './AdminModal';

type RemoveAdminUserModalProps = {
  open: boolean;
  setOpen(opened: boolean): void;
  sciper: number;
  handleRemoveRoleUser(): void;
};

const RemoveAdminUserModal: FC<RemoveAdminUserModalProps> = ({
  open,
  setOpen,
  sciper,
  handleRemoveRoleUser,
}) => {
  const { t } = useTranslation();
  const fctx = useContext(FlashContext);

  const [loading, setLoading] = useState(false);

  const handleCancel = () => setOpen(false);

  const handleDelete = async () => {
    const requestOptions = {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ sciper: sciper }),
    };

    try {
      setLoading(true);
      const res = await fetch(ENDPOINT_REMOVE_ROLE, requestOptions);
      if (res.status !== 200) {
        const response = await res.text();
        fctx.addMessage(
          `Error HTTP ${res.status} (${res.statusText}) : ${response}`,
          FlashLevel.Error
        );
      } else {
        handleRemoveRoleUser();
        fctx.addMessage(t('successRemoveUser'), FlashLevel.Info);
      }
    } catch (error) {
      fctx.addMessage(`${t('errorRemoveUser')}: ${error.message}`, FlashLevel.Error);
    }
    setLoading(false);
    setOpen(false);
  };
  const cancelButtonRef = useRef(null);

  const modalBody = (
    <Dialog.Title as="h3" className="text-lg leading-6 font-medium text-gray-900">
      {t('confirmDeleteUserSciper')} {sciper}
    </Dialog.Title>
  );

  const actionButton = (
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

RemoveAdminUserModal.propTypes = {
  open: PropTypes.bool.isRequired,
  setOpen: PropTypes.func.isRequired,
  sciper: PropTypes.number.isRequired,
};

export default RemoveAdminUserModal;
