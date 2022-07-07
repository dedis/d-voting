import React, { FC } from 'react';
import PropTypes from 'prop-types';

import { ID } from 'types/configuration';
import { OngoingAction, Status } from 'types/election';
import { NodeStatus } from 'types/node';
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
  nodeToSetup,
  setNodeToSetup,
  DKGStatuses,
  setDKGStatuses,
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
    setNodeToSetup,
    DKGStatuses,
    setDKGStatuses
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
