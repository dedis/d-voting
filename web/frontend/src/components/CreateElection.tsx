import React, { FC, useState } from 'react';
import { useTranslation } from 'react-i18next';

import ElectionForm from './ElectionForm';
import Modal from './modal/Modal';

const CreateElection: FC = () => {
  const { t } = useTranslation();
  const [showModal, setShowModal] = useState(false);
  const [textModal, setTextModal] = useState('');

  return (
    <div className="create-election-wrapper">
      <Modal
        showModal={showModal}
        setShowModal={setShowModal}
        textModal={textModal}
        buttonRightText={t('close')}
      />
      <h1 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
        {t('navBarCreate')}
      </h1>
      <p className="py-5">{t('create')}</p>

      <div className="election-form flex flex-row justify-center items-start">
        <ElectionForm setShowModal={setShowModal} setTextModal={setTextModal} />
      </div>
    </div>
  );
};

export default CreateElection;
