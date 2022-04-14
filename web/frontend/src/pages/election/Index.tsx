import React, { FC } from 'react';
import { useTranslation } from 'react-i18next';

import ElectionTable from './components/ElectionTable';
import useFetchCall from 'components/utils/useFetchCall';
import { ENDPOINT_EVOTING_GET_ALL } from 'components/utils/Endpoints';
import './Index.css';

const ElectionIndex: FC = () => {
  const { t } = useTranslation();
  const request = {
    method: 'GET',
  };

  const [data, loading, error] = useFetchCall(ENDPOINT_EVOTING_GET_ALL, request);

  /*Show all the elections retrieved if any */
  const showElection = () => {
    return (
      <div>
        {data.Elections.length > 0 ? (
          <div className="election-box">
            <div className="click-info">{t('clickElection')}</div>
            <div className="election-table-wrapper">
              <ElectionTable elections={data.Elections} />
            </div>
          </div>
        ) : (
          <div className="no-election">{t('noElection')}</div>
        )}
      </div>
    );
  };

  return (
    <div className="election-wrapper">
      {t('listElection')}
      {!loading ? (
        showElection()
      ) : error === null ? (
        <p className="loading">{t('loading')} </p>
      ) : (
        <div className="error-retrieving">{t('errorRetrievingElection')}</div>
      )}
    </div>
  );
};

export default ElectionIndex;
