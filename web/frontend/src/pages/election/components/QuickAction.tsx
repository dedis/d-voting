import React, { FC } from 'react';
import { ID } from 'types/configuration';
import { Status } from 'types/election';
import ResultButton from './ActionButtons/ResultButton';
import VoteButton from './ActionButtons/VoteButton';

type QuickActionProps = {
  status: Status;
  electionID: ID;
};

// TODO fetch the results
const QuickAction: FC<QuickActionProps> = ({ status, electionID }) => {
  return (
    <div>
      {status === Status.Open && <VoteButton status={status} electionID={electionID} />}
      {status === Status.ResultAvailable && (
        <ResultButton status={status} electionID={electionID} />
      )}
      {status !== Status.Open && status !== Status.ResultAvailable && <span></span>}
    </div>
  );
};

export default QuickAction;
