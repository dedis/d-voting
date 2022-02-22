import React, { FC } from 'react';
import { useTranslation } from 'react-i18next';

import { ROUTE_RESULT_SHOW } from '../../Routes';
import SimpleTable from '../../components/utils/SimpleTable';
import { RESULT_AVAILABLE } from '../../components/utils/StatusNumber';

const ResultIndex: FC = () => {
  const { t } = useTranslation();
  return (
    <div>
      <SimpleTable
        statusToKeep={RESULT_AVAILABLE}
        pathLink={ROUTE_RESULT_SHOW}
        textWhenData={t('displayResults')}
        textWhenNoData={t('noResultsAvailable')}
      />
    </div>
  );
};

export default ResultIndex;
