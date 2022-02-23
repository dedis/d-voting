import React, { FC, useState } from 'react';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

import useChangeAction from 'components/utils/useChangeAction';
import Modal from 'components/modal/Modal';
import './Status.css';

type ActionProps = {
  status: number;
  electionID: number;
  setStatus: (status: number) => void;
  setResultAvailable?: (available: boolean) => void | null;
};

/**/
const Action: FC<ActionProps> = ({ status, electionID, setStatus, setResultAvailable }) => {
  const { t } = useTranslation();

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

  return (
    <span>
      {getAction()}
      {modalClose}
      {modalCancel}
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
  electionID: PropTypes.number.isRequired,
  setStatus: PropTypes.func.isRequired,
  setResultAvailable: PropTypes.func,
};

export default Action;
