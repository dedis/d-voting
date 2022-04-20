import React, { FC } from 'react';
import { Link } from 'react-router-dom';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

import { ROUTE_RESULT_INDEX } from 'Routes';
import useElection from 'components/utils/useElection';
import Result from './components/Result';
import { useConfigurationOnly } from 'components/utils/useConfiguration';
import BackButton from './components/BackButton';

type ResultShowProps = {
  location?: any;
};

const ResultShow: FC<ResultShowProps> = (props) => {
  const { t } = useTranslation();
  //props.location.data = id of the election
  const { loading, configObj, result, error } = useElection(props.location.data);
  const configuration = useConfigurationOnly(configObj);

  return (
    <div>
      {!loading ? (
        <div className="shadow-lg rounded-md w-full my-0 sm:my-4">
          <h1 className="px-4 text-2xl text-gray-900 sm:text-3xl sm:truncate">
            <span className="font-bold">Results: </span>
            {configuration.MainTitle}
          </h1>
          {<Result resultData={result} configuration={configuration} />}
        </div>
      ) : error === null ? (
        <p className="loading">{t('loading')} </p>
      ) : (
        <div className="error-retrieving">{t('errorRetrievingElection')}</div>
      )}
      <Link to={ROUTE_RESULT_INDEX}>
        <BackButton />
      </Link>
    </div>
  );
};

ResultShow.propTypes = {
  location: PropTypes.any,
};

export default ResultShow;
