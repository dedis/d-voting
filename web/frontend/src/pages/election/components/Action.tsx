import React, { FC } from 'react';
import PropTypes from 'prop-types';

import { ID } from 'types/configuration';
import { OngoingAction, Status } from 'types/election';
import useChangeAction from './utils/useChangeAction';

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
  nodeToSetup: [string, string];
  setNodeToSetup: ([node, proxy]: [string, string]) => void;
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
  nodeToSetup,
  setNodeToSetup,
}) => {
  const { getAction, modalClose, modalCancel, modalDelete, modalSetup } = useChangeAction(
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
    nodeToSetup,
    setNodeToSetup
  );

  return (
    <>
      {getAction()}
      {modalClose}
      {modalCancel}
      {modalDelete}
      {modalSetup}
    </>
  );
};

Action.propTypes = {
  status: PropTypes.number.isRequired,
  electionID: PropTypes.string.isRequired,
  setStatus: PropTypes.func.isRequired,
  setResultAvailable: PropTypes.func,
};

export default Action;
