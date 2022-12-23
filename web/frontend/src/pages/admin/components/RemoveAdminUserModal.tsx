import React, { FC, useContext, useEffect, useState } from 'react';
import { ENDPOINT_REMOVE_ROLE } from 'components/utils/Endpoints';
import PropTypes from 'prop-types';
import { Dialog } from '@headlessui/react';
import { UserRemoveIcon } from '@heroicons/react/outline';
import SpinnerIcon from 'components/utils/SpinnerIcon';
import { useTranslation } from 'react-i18next';
import { FlashContext, FlashLevel } from 'index';
import AdminModal from './AdminModal';
import usePostCall from 'components/utils/usePostCall';

type RemoveAdminUserModalProps = {
  open: boolean;
  setOpen(opened: boolean): void;
  sciper: number;
  handleRemoveRoleUser(user: object): void;
};

const RemoveAdminUserModal: FC<RemoveAdminUserModalProps> = ({
  open,
  setOpen,
  sciper,
  handleRemoveRoleUser,
}) => {
  const { t } = useTranslation();
  const fctx = useContext(FlashContext);
  const [postError, setPostError] = useState(null);
  const [, setIsPosting] = useState(false);
  const [loading, setLoading] = useState(false);

  const handleCancel = () => {
    setOpen(false);
  };
  const sendFetchRequest = usePostCall(setPostError);

  useEffect(() => {
    if (postError !== null) {
      fctx.addMessage(t('addRoleError') + postError, FlashLevel.Error);
      setPostError(null);
    }
  }, [postError]);
  const usersToBeRemoved = [sciper];
  const saveMapping = async () => {
    const request = {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(usersToBeRemoved),
    };
    return sendFetchRequest(ENDPOINT_REMOVE_ROLE, request, setIsPosting);
  };
  const handleDelete = async () => {
    setLoading(true);
    if (sciper !== 0) {
      try {
        const res = await saveMapping();
        if (!res) {
          handleRemoveRoleUser(usersToBeRemoved);
          fctx.addMessage(t('successRemoveUser'), FlashLevel.Info);
        }
        setOpen(false);
      } catch {
        fctx.addMessage(t('errorRemoveUser'), FlashLevel.Error);
      }
    } else {
        fctx.addMessage(t('errorRemoveUser'), FlashLevel.Error);
    }

    setLoading(false);
  };

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
