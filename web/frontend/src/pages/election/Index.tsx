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
    return data.Elections.length > 0 ? (
      <>
        <div className="pb-8">
          <h2 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
            {t('elections')}
          </h2>
          <div>{t('listElection')}</div>
          <div>{t('clickElection')}</div>
        </div>
        <div>
          <ElectionTable elections={data.Elections} />
        </div>
      </>
    ) : (
      <div className="no-election">{t('noElection')}</div>
    );
  };

  return (
    <div className="w-[60rem] font-sans px-4 py-4">
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
