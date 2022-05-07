import React, { FC, useState } from 'react';
import { useTranslation } from 'react-i18next';

import ElectionForm from './components/ElectionForm';
import Modal from 'components/modal/Modal';

const ElectionCreate: FC = () => {
  const { t } = useTranslation();
  const [showModal, setShowModal] = useState(false);
  const [textModal, setTextModal] = useState('');

  return (
    <div className="font-sans">
      <Modal
        showModal={showModal}
        setShowModal={setShowModal}
        textModal={textModal}
        buttonRightText={t('close')}
      />
      <div className="px-4 py-4">
        <h2 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
          {t('navBarCreate')}
        </h2>
        <div className="pt-2">{t('create')}</div>
      </div>

      <ElectionForm setShowModal={setShowModal} setTextModal={setTextModal} />
    </div>
  );
};

export default ElectionCreate;
