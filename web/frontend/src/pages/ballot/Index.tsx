import React, { FC } from 'react';
import { useTranslation } from 'react-i18next';

import SimpleTable from 'components/utils/SimpleTable';
import { OPEN } from 'components/utils/StatusNumber';
import './Index.css';

const BallotIndex: FC = () => {
  const { t } = useTranslation();

  return (
    <div>
      <SimpleTable
        statusToKeep={OPEN}
        pathLink="vote"
        textWhenData={t('voteAllowed')}
        textWhenNoData={t('noVote')}
      />
    </div>
  );
};

export default BallotIndex;
