import { NoActionAvailable, ResultButton, VoteButton } from 'components/utils/ActionButtons';
import React, { FC } from 'react';
import { ID } from 'types/configuration';
import { Status } from 'types/election';

type QuickActionProps = {
  status: Status;
  electionID: ID;
};

const QuickAction: FC<QuickActionProps> = ({ status, electionID }) => {
  return (
    <div>
      {status == Status.Open && <VoteButton status={status} electionID={electionID} />}
      {status == Status.ResultAvailable && <ResultButton status={status} electionID={electionID} />}
      {status !== Status.Open && status !== Status.ResultAvailable && <NoActionAvailable />}
    </div>
  );
};

export default QuickAction;
