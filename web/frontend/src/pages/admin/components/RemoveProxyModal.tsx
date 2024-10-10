import React, { FC, useContext, useEffect, useState } from 'react';
import { MinusCircleIcon } from '@heroicons/react/outline';
import { Dialog } from '@headlessui/react';
import SpinnerIcon from 'components/utils/SpinnerIcon';
import { FlashContext, FlashLevel } from 'index';
import { useTranslation } from 'react-i18next';
import usePostCall from 'components/utils/usePostCall';
import * as endpoints from 'components/utils/Endpoints';
import AdminModal from './AdminModal';

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
  const [, setIsPosting] = useState(false);

  const sendFetchRequest = usePostCall(setPostError);

  useEffect(() => {
    if (postError !== null) {
      fctx.addMessage(t('removeProxyError') + postError, FlashLevel.Error);
      setPostError(null);
    }
  }, [fctx, t, postError]);

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

  const handleCancel = () => setOpen(false);

  const modalBody = (
    <>
      <Dialog.Title as="h3" className="text-lg leading-6 font-medium text-gray-900">
        {t('confirmDeleteProxy')}
      </Dialog.Title>
      <div className="my-4">
        {t('node')}: {node}
      </div>
    </>
  );

  const actionButton = (
    <button
      type="button"
      className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-[#ff0000] text-base font-medium text-white hover:bg-[#b51f1f] focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-[#ff0000] sm:col-start-2 sm:text-sm"
      onClick={handleDelete}>
      {loading ? (
        <SpinnerIcon />
      ) : (
        <MinusCircleIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
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

export default RemoveProxyModal;
