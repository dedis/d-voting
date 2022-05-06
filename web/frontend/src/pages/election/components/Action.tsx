import React, { FC, useContext, useState } from 'react';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

import Modal from 'components/modal/Modal';
import { ID } from 'types/configuration';
import useChangeAction from 'components/utils/useChangeAction';
import { STATUS } from 'types/election';
import DeleteButton from 'components/utils/DeleteButton';
import { FlashContext, FlashLevel } from 'index';
import { useNavigate } from 'react-router-dom';

type ActionProps = {
  status: STATUS;
  electionID: ID;
  setStatus: (status: STATUS) => void;
  setResultAvailable?: (available: boolean) => void | null;
};

const Action: FC<ActionProps> = ({ status, electionID, setStatus, setResultAvailable }) => {
  const { t } = useTranslation();
  const fctx = useContext(FlashContext);
  const navigate = useNavigate();

  const [textModalError, setTextModalError] = useState(null);
  const [showModalError, setShowModalError] = useState(false);
  const { getAction, modalClose, modalCancel } = useChangeAction(
    status,
    electionID,
    setStatus,
    setResultAvailable,
    setTextModalError,
    setShowModalError
  );

  const deleteElection = async () => {
    const request = {
      method: 'DELETE',
    };

    const res = await fetch(`/api/evoting/elections/${electionID}`, request);
    if (!res.ok) {
      const txt = await res.text();
      fctx.addMessage(`failed to send delete request: ${txt}`, FlashLevel.Error);
      return;
    }

    fctx.addMessage('election deleted', FlashLevel.Info);
    navigate('/');
  };

  return (
    <span>
      {getAction()}
      {modalClose}
      {modalCancel}
      <DeleteButton status={status} handleDelete={deleteElection} />
      {
        <Modal
          showModal={showModalError}
          setShowModal={setShowModalError}
          textModal={textModalError === null ? '' : textModalError}
          buttonRightText={t('close')}
        />
      }
    </span>
  );
};

Action.propTypes = {
  status: PropTypes.number.isRequired,
  electionID: PropTypes.string.isRequired,
  setStatus: PropTypes.func.isRequired,
  setResultAvailable: PropTypes.func,
};

export default Action;
