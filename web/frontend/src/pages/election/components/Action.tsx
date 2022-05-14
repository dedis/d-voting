import React, { FC, useContext } from 'react';
import PropTypes from 'prop-types';

import { ID } from 'types/configuration';
import { OngoingAction, Status } from 'types/election';
import useChangeAction from 'components/utils/useChangeAction';
import { NodeStatus } from 'types/node';
import DeleteButton from 'components/buttons/DeleteButton';
import { FlashContext, FlashLevel } from 'index';
import { useNavigate } from 'react-router-dom';

type ActionProps = {
  status: Status;
  electionID: ID;
  roster: string[];
  nodeProxyAddresses: Map<string, string>;
  setStatus: (status: Status) => void;
  setResultAvailable?: (available: boolean) => void | null;
  setTextModalError: (text: string) => void;
  setShowModalError: (show: boolean) => void;
  ongoingAction: OngoingAction;
  setOngoingAction: (action: OngoingAction) => void;
  DKGStatuses: Map<string, NodeStatus>;
  setDKGStatuses: (dkgStatuses: Map<string, NodeStatus>) => void;
};

const Action: FC<ActionProps> = ({
  status,
  electionID,
  roster,
  nodeProxyAddresses,
  setStatus,
  setResultAvailable,
  setTextModalError,
  setShowModalError,
  ongoingAction,
  setOngoingAction,
  DKGStatuses,
  setDKGStatuses,
}) => {
  const fctx = useContext(FlashContext);
  const navigate = useNavigate();
  const { getAction, modalClose, modalCancel } = useChangeAction(
    status,
    electionID,
    roster,
    nodeProxyAddresses,
    setStatus,
    setResultAvailable,
    setTextModalError,
    setShowModalError,
    ongoingAction,
    setOngoingAction,
    DKGStatuses,
    setDKGStatuses
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
