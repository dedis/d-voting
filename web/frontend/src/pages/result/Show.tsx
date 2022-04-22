import React, { FC } from 'react';
import { Link } from 'react-router-dom';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

import { ROUTE_RESULT_INDEX } from 'Routes';
import useElection from 'components/utils/useElection';
import Result from './components/Result';
import { useConfigurationOnly } from 'components/utils/useConfiguration';
import TextButton from '../../components/buttons/TextButton';

type ResultShowProps = {
  location?: any;
};

// Is this page really necessary ?
const ResultShow: FC<ResultShowProps> = (props) => {
  const { t } = useTranslation();
  //props.location.data = id of the election
  const { loading, configObj, result, error } = useElection(props.location.data);
  const configuration = useConfigurationOnly(configObj);

  return (
    <div>
      {!loading ? (
        <div>
          <h1 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
            Results
          </h1>
          {<Result resultData={result} configuration={configuration} />}
        </div>
      ) : error === null ? (
        <div>
          <p className="loading">{t('loading')} </p>
          <Link to={ROUTE_RESULT_INDEX}>
            <TextButton>{t('back')}</TextButton>
          </Link>
        </div>
      ) : (
        <div>
          <div className="error-retrieving">{t('errorRetrievingElection')}</div>
          <Link to={ROUTE_RESULT_INDEX}>
            <TextButton>{t('back')}</TextButton>
          </Link>
        </div>
      )}
    </div>
  );
};

ResultShow.propTypes = {
  location: PropTypes.any,
};

export default ResultShow;
