import React, { FC } from 'react';
import { useTranslation } from 'react-i18next';

import ElectionTable from './components/ElectionTable';
import useFetchCall from 'components/utils/useFetchCall';
import * as endpoints from 'components/utils/Endpoints';
import './Index.css';
import Loading from 'pages/Loading';

const ElectionIndex: FC = () => {
  const { t } = useTranslation();
  const request = {
    method: 'GET',
    headers: {
      'Access-Control-Allow-Origin': '*',
    },
  };

  const [data, loading, error] = useFetchCall(endpoints.elections, request);

  /*Show all the elections retrieved if any */
  const showElection = () => {
    return (
      <div>
        {data.Elections.length > 0 ? (
          <>
            {t('listElection')}
            <div className="pb-8">{t('clickElection')}</div>
            <div>
              <ElectionTable elections={data.Elections} />
            </div>
          </>
        ) : (
          <div className="no-election">{t('noElection')}</div>
        )}
      </div>
    );
  };

  return (
    <div className="pt-4 mx-2">
      {!loading ? (
        showElection()
      ) : error === null ? (
        <Loading />
      ) : (
        <div className="error-retrieving">
          {t('errorRetrievingElection')} - {error.toString()}
        </div>
      )}
    </div>
  );
};

export default ElectionIndex;
