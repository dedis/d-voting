import React, { FC } from 'react';
import { ID } from 'types/configuration';
import { Status } from 'types/form';
import ResultButton from './ActionButtons/ResultButton';
import VoteButton from './ActionButtons/VoteButton';

type QuickActionProps = {
  status: Status;
  formID: ID;
};

const QuickAction: FC<QuickActionProps> = ({ status, formID }) => {
  return (
    <div>
      {status === Status.Open && <VoteButton status={status} formID={formID} />}
      {status === Status.ResultAvailable && <ResultButton status={status} formID={formID} />}
    </div>
  );
};

export default QuickAction;
