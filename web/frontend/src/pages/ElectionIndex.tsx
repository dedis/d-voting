import React, { FC } from 'react';
import { useTranslation } from 'react-i18next';

import ElectionTable from '../components/ElectionTable';
import useFetchCall from '../components/utils/useFetchCall';
import { ENDPOINT_EVOTING_GET_ALL } from '../components/utils/Endpoints';
import './Election.css';

/*Assumption : for now an election is simply a json file with the following field
    - electionName: string
    - Format: []byte -> it stores the election questions 
    - electionStatus : number
    - collectivePublicKey :
    - electionID : string
*/
/*Disclaimer : 
Currently the Format parameter of an election is always a []string
called Candidates
 */

const ElectionIndex: FC = () => {
  const { t } = useTranslation();
  const token = sessionStorage.getItem('token');
  const request = {
    method: 'POST',
    body: JSON.stringify({ Token: token }),
  };
  const [data, loading, error] = useFetchCall(ENDPOINT_EVOTING_GET_ALL, request);

  /*Show all the elections retrieved if any */
  const showElection = () => {
    return (
      <div>
        {data.AllElectionsInfo.length > 0 ? (
          <div className="election-box">
            <div className="click-info">{t('clickElection')}</div>
            <div className="election-table-wrapper">
              <ElectionTable elections={data.AllElectionsInfo} />
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
