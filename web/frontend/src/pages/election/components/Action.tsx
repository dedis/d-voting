import React, { FC } from 'react';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

//import Modal from 'components/modal/Modal';
import { ID } from 'types/configuration';
import { OngoingAction, Status } from 'types/election';
import useChangeAction from 'components/utils/useChangeAction';

type ActionProps = {
  status: Status;
  electionID: ID;
  roster: string[];
  setStatus: (status: Status) => void;
  setResultAvailable?: (available: boolean) => void | null;
  setGetError: (error: string) => void;
  setTextModalError: (text: string) => void;
  setShowModalError: (show: boolean) => void;
  ongoingAction: OngoingAction;
  setOngoingAction: (action: OngoingAction) => void;
};

const Action: FC<ActionProps> = ({
  status,
  electionID,
  roster,
  setStatus,
  setResultAvailable,
  setGetError,
  setTextModalError,
  setShowModalError,
  ongoingAction,
  setOngoingAction,
}) => {
  const { t } = useTranslation();

  const { getAction, modalClose, modalCancel, modalAddProxyAddresses } = useChangeAction(
    status,
    electionID,
    roster,
    setStatus,
    setResultAvailable,
    setTextModalError,
    setShowModalError,
    setGetError,
    ongoingAction,
    setOngoingAction
  );

  return (
    <span>
      {getAction()}
      {modalClose}
      {modalCancel}
      {modalAddProxyAddresses}
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
