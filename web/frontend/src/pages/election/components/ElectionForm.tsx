import React, { FC } from 'react';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

import { ENDPOINT_EVOTING_CREATE } from 'components/utils/Endpoints';

type ElectionFormProps = {
  setShowModal(modal: any): void;
  setTextModal(text: string): void;
};

const ElectionForm: FC<ElectionFormProps> = ({ setShowModal, setTextModal }) => {
  const { t } = useTranslation();

  async function createHandler() {
    const data = {
      Title: '',
      AdminID: '',
      Token: '',
      Format: '',
    };

    const req = {
      method: 'POST',
      type: 'application/json',
      body: JSON.stringify(data),
    };
    const res = await fetch(ENDPOINT_EVOTING_CREATE, req);
    if (res.status !== 200) {
      console.warn('failed to create election');
    }
  }

  return (
    <div className="form-wrapper bg-gray-200 flex-1 m-1 p-10">
      <button onClick={createHandler}>{t('createElection')}</button>
    </div>
  );
};

ElectionForm.propTypes = {
  setShowModal: PropTypes.func.isRequired,
  setTextModal: PropTypes.func.isRequired,
};

export default ElectionForm;
