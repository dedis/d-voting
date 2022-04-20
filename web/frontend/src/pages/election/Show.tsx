import React, { FC, useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

import Action from './components/Action';
import Status from './components/Status';
import useElection from 'components/utils/useElection';
import { RESULT_AVAILABLE } from 'components/utils/StatusNumber';
import { ROUTE_ELECTION_INDEX } from 'Routes';
import './Show.css';
import useGetResults from 'components/utils/useGetResults';
import Result from 'pages/result/components/Result';
import { useConfigurationOnly } from 'components/utils/useConfiguration';
import BackButton from 'pages/result/components/BackButton';

const ElectionShow: FC = () => {
  const { t } = useTranslation();
  const { electionId } = useParams();

  const {
    loading,
    electionID,
    status,
    setStatus,
    pubKey,
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
  //fetch result when available after a status change
  useEffect(() => {
    if (status === RESULT_AVAILABLE && isResultAvailable) {
      getResults(electionID, setError, setResult, setIsResultSet);
    }
  }, [isResultAvailable, status]);

  return (
    <div>
      {!loading ? (
        <div className="shadow-lg rounded-md w-full my-0 sm:my-4">
          <h1 className="px-4 text-2xl text-gray-900 sm:text-3xl sm:truncate">
            <span className="font-bold">Results: </span>
            {configuration.MainTitle}
          </h1>

          {isResultSet ? (
            <Result resultData={result} configuration={configuration} />
          ) : (
            <div className="px-4 pb-4">
              {t('status')}:<Status status={status} />
              <span>
                Action :
                <Action
                  status={status}
                  electionID={electionID}
                  setStatus={setStatus}
                  setResultAvailable={setIsResultAvailable}
                />{' '}
              </span>
              <div>
                {/* TODO: Maybe replace with a button to go vote in the election
                  if available
                  candidates.map((cand) => (
                    <li key={cand} className="election-candidate">
                      {cand}
                    </li>
                  ))*/}
              </div>
            </div>
          )}

          <Link to={ROUTE_ELECTION_INDEX}>
            <BackButton />
          </Link>
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
