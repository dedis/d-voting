import React, { FC } from 'react';
import { Link } from 'react-router-dom';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

import { ROUTE_RESULT_INDEX } from 'Routes';
import useElection from 'components/utils/useElection';
import './Show.css';

type ResultShowProps = {
  location?: any;
};

const ResultShow: FC<ResultShowProps> = (props) => {
  const { t } = useTranslation();
  //props.location.data = id of the election
  const { loading, configObj, result, error } = useElection(props.location.data);

  return (
    <div className="result-box">
      {!loading ? (
        <div>
          <h1>{configObj.MainTitle}</h1>
          {/* TODO: <Result resultData={result} candidates={candidates} />*/}
        </div>
      ) : error === null ? (
        <p className="loading">{t('loading')} </p>
      ) : (
        <div className="error-retrieving">{t('errorRetrievingElection')}</div>
      )}
      <Link to={ROUTE_RESULT_INDEX}>
        <button className="back-btn">{t('back')}</button>
      </Link>
    </div>
  );
};

ResultShow.propTypes = {
  location: PropTypes.any,
};

export default ResultShow;
