import React, { FC, useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

import { CREATE_ENDPOINT } from '../utils/Endpoints';
import usePostCall from '../utils/usePostCall';
import {
  COLLECTIVE_AUTHORITY_MEMBERS,
  SHUFFLE_THRESHOLD,
} from '../utils/CollectiveAuthorityMembers';

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
    const res = await fetch(CREATE_ENDPOINT, req);
    if (res.status !== 200) {
      console.warn('failed to create election');
    }
  }

  return (
    <div className="form-wrapper bg-gray-200 flex-1 m-1 p-10">
      <button onClick={createHandler}>Create election</button>
    </div>
  );
};

ElectionForm.propTypes = {
  setShowModal: PropTypes.func.isRequired,
  setTextModal: PropTypes.func.isRequired,
};

export default ElectionForm;
