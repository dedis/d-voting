import React, { FC, useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

import Action from './components/Action';
import Status from './components/Status';
import useElection from 'components/utils/useElection';
import { ROUTE_ELECTION_INDEX } from 'Routes';
import './Show.css';
import useGetResults from 'components/utils/useGetResults';
import { STATUS } from 'types/electionInfo';

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
    chunksPerBallot,
    ballotSize,
    configObj,
    isResultSet,
    setIsResultSet,
    error,
  } = useElection(electionId);

  const [, setError] = useState(null);
  const [isResultAvailable, setIsResultAvailable] = useState(false);
  const { getResults } = useGetResults();
  //fetch result when available after a status change
  useEffect(() => {
    if (status === STATUS.RESULT_AVAILABLE && isResultAvailable) {
      getResults(electionID, setError, setResult, setIsResultSet);
    }
  }, [isResultAvailable, status]);

  return (
    <div className="election-details-box">
      {!loading ? (
        <div>
          <h1>{configObj.MainTitle}</h1>
          <div className="election-details-wrapper">
            {isResultSet ? (
              <div className="election-wrapper-child">
                {/* TODO: <Result resultData={result} candidates={candidates} />*/}
              </div>
            ) : (
              <div className="election-wrapper-child">
                {' '}
                {t('status')}:<Status status={status} />
                <span className="election-action">
                  Action :
                  <Action
                    status={status}
                    electionID={electionID}
                    setStatus={setStatus}
                    setResultAvailable={setIsResultAvailable}
                  />{' '}
                </span>
                <div className="election-candidates">
                  {t('candidates')}
                  {/* TODO: candidates.map((cand) => (
                    <li key={cand} className="election-candidate">
                      {cand}
                    </li>
                  ))*/}
                </div>
              </div>
            )}

            <Link to={ROUTE_ELECTION_INDEX}>
              <button className="back-btn">{t('back')}</button>
            </Link>
            {/* <Link to={ROUTE_RESULT_INDEX}>
              <button className="back-btn">{t('back')}</button>
            </Link> */}
          </div>
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
