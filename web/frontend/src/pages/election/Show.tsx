import React, { FC, useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

import useElection from 'components/utils/useElection';
import './Show.css';
import useGetResults from 'components/utils/useGetResults';
import Result from 'pages/result/components/Result';
import { useConfigurationOnly } from 'components/utils/useConfiguration';
import { STATUS } from 'types/electionInfo';
import ResultNotAvailable from '../result/components/ResultNotAvailable';

const ElectionShow: FC = () => {
  const { t } = useTranslation();
  const { electionId } = useParams();

  const {
    loading,
    electionID,
    status,
    setStatus,
    result,
    setResult,
    configObj,
    isResultSet,
    setIsResultSet,
    error,
  } = useElection(electionId);

  const configuration = useConfigurationOnly(configObj);
  const [, setError] = useState(null);
  const [isResultAvailable, setIsResultAvailable] = useState(false);
  const { getResults } = useGetResults();

  //Fetch result when available after a status change
  useEffect(() => {
    if (status === STATUS.RESULT_AVAILABLE && isResultAvailable) {
      getResults(electionID, setError, setResult, setIsResultSet);
    }
  }, [isResultAvailable, status]);

  return (
    <div>
      {!loading ? (
        <div>
          <h1 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
            Results
          </h1>
          {isResultSet ? (
            <Result resultData={result} configuration={configuration} />
          ) : (
            <ResultNotAvailable
              status={status}
              setStatus={setStatus}
              setIsResultAvailable={setIsResultAvailable}
              configuration={configuration}
              electionID={electionID}
            />
          )}
        </div>
      ) : (
        <p className="loading">{t('loading')}</p>
      )}
    </div>
  );
};

ElectionShow.propTypes = {
  location: PropTypes.any,
};

export default ElectionShow;
