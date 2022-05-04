import { CastVoteButton, ResultButton } from 'components/utils/ActionButtons';
import React, { FC } from 'react';
import { ID } from 'types/configuration';
import { STATUS } from 'types/election';

type QuickActionProps = {
  status: STATUS;
  electionID: ID;
};

const QuickAction: FC<QuickActionProps> = ({ status, electionID }) => {
  return (
    <div>
      {status == STATUS.Open && <CastVoteButton status={status} electionID={electionID} />}
      {status == STATUS.ResultAvailable && <ResultButton status={status} electionID={electionID} />}
    </div>
  );
};

export default QuickAction;
