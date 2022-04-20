import React, { FC } from 'react';
import { useTranslation } from 'react-i18next';

import { ROUTE_RESULT_SHOW } from 'Routes';
import SimpleTable from 'pages/election/components/SimpleTable';
import { STATUS } from 'types/electionInfo';

const ResultIndex: FC = () => {
  const { t } = useTranslation();
  return (
    <div>
      <SimpleTable
        statusToKeep={STATUS.RESULT_AVAILABLE}
        pathLink={ROUTE_RESULT_SHOW}
        textWhenData={t('displayResults')}
        textWhenNoData={t('noResultsAvailable')}
      />
    </div>
  );
};

export default ResultIndex;
