import React, { FC } from 'react';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

//import Modal from 'components/modal/Modal';
import { ID } from 'types/configuration';
import useChangeAction from 'components/utils/useChangeAction';
import { STATUS } from 'types/election';

type ActionProps = {
  status: STATUS;
  electionID: ID;
  nodeRoster: string[];
  setStatus: (status: STATUS) => void;
  setResultAvailable?: (available: boolean) => void | null;
  setGetError: (error: string) => void;
  setTextModalError: (text: string) => void;
  setShowModalError: (show: boolean) => void;
};

const Action: FC<ActionProps> = ({
  status,
  electionID,
  nodeRoster,
  setStatus,
  setResultAvailable,
  setGetError,
  setTextModalError,
  setShowModalError,
}) => {
  const { t } = useTranslation();

  const { getAction, modalClose, modalCancel } = useChangeAction(
    status,
    electionID,
    nodeRoster,
    setStatus,
    setResultAvailable,
    setTextModalError,
    setShowModalError,
    setGetError
  );

  return (
    <span>
      {getAction()}
      {modalClose}
      {modalCancel}
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
