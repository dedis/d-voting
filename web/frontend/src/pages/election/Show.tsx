import React, { FC, useContext, useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

import useElection from 'components/utils/useElection';
import './Show.css';
import useGetResults from 'components/utils/useGetResults';
import { STATUS } from 'types/electionInfo';
import Status from './components/Status';
import Action from './components/Action';
import { ROUTE_BALLOT_SHOW, ROUTE_ELECTION_INDEX } from 'Routes';
import TextButton from 'components/buttons/TextButton';
import { AuthContext } from 'index';
import { Role } from 'types/userRole';

const ElectionShow: FC = () => {
  const { t } = useTranslation();
  const { electionId } = useParams();
  const authCtx = useContext(AuthContext);

  const { loading, electionID, status, setStatus, setResult, configObj, setIsResultSet } =
    useElection(electionId);

  const [, setError] = useState(null);
  const [isResultAvailable, setIsResultAvailable] = useState(false);
  const { getResults } = useGetResults();

  //Fetch result when available after a status change
  useEffect(() => {
    if (status === STATUS.ResultAvailable && isResultAvailable) {
      getResults(electionID, setError, setResult, setIsResultSet);
    }
  }, [isResultAvailable, status]);

  return (
    <div>
      {!loading ? (
        <div>
          <div className="shadow-lg rounded-md w-full px-4 my-0 sm:my-4">
            <h3 className="py-6 uppercase text-2xl text-center text-gray-700">
              {configObj.MainTitle}
            </h3>
            <div className="px-4">
              {t('status')}: <Status status={status} />
              <span className="mx-4">{t('action')}:</span>
              <Action
                status={status}
                electionID={electionID}
                setStatus={setStatus}
                setResultAvailable={setIsResultAvailable}
              />
            </div>
          </div>
          <div className="flex my-4">
            {status === STATUS.Open &&
            authCtx.isLogged &&
            (authCtx.role === Role.Admin ||
              authCtx.role === Role.Operator ||
              authCtx.role === Role.Voter) ? (
              <Link to={ROUTE_BALLOT_SHOW + '/' + electionID}>
                <TextButton>{t('navBarVote')}</TextButton>
              </Link>
            ) : null}
            <Link to={ROUTE_ELECTION_INDEX}>
              <TextButton>{t('back')}</TextButton>
            </Link>
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
