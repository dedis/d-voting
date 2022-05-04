import React, { FC, useContext, useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

import useElection from 'components/utils/useElection';
import './Show.css';
import useGetResults from 'components/utils/useGetResults';
import { STATUS } from 'types/election';
import Status from './components/Status';
import Action from './components/Action';
import { ROUTE_BALLOT_SHOW, ROUTE_ELECTION_INDEX } from 'Routes';
import TextButton from 'components/buttons/TextButton';
import { AuthContext } from 'index';
import { ROLE } from 'types/userRole';
import Modal from 'components/modal/Modal';

const ElectionShow: FC = () => {
  const { t } = useTranslation();
  const { electionId } = useParams();
  const authCtx = useContext(AuthContext);

  const { loading, electionID, status, setStatus, roster, setResult, configObj, setIsResultSet } =
    useElection(electionId);

  const [, setError] = useState(null);
  const [isResultAvailable, setIsResultAvailable] = useState(false);
  const { getResults } = useGetResults();

  const [getError, setGetError] = useState(null);
  const [textModalError, setTextModalError] = useState(null);
  const [showModalError, setShowModalError] = useState(false);

  //Fetch result when available after a status change
  useEffect(() => {
    if (status === STATUS.ResultAvailable && isResultAvailable) {
      getResults(electionID, setError, setResult, setIsResultSet);
    }
  }, [isResultAvailable, status]);

  useEffect(() => {
    if (getError !== null) {
      console.log(getError);
      setTextModalError(getError);
      setShowModalError(true);
      setGetError(null);
    }
  }, [getError, setTextModalError]);

  return (
    <div>
      {!loading ? (
        <div>
          <Modal
            showModal={showModalError}
            setShowModal={setShowModalError}
            textModal={textModalError === null ? '' : textModalError}
            buttonRightText={t('close')}
          />
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
                nodeRoster={roster}
                setStatus={setStatus}
                setResultAvailable={setIsResultAvailable}
                setGetError={setGetError}
                setTextModalError={setTextModalError}
                setShowModalError={setShowModalError}
              />
            </div>
          </div>
          <div className="flex my-4">
            {status === STATUS.Open &&
            authCtx.isLogged &&
            (authCtx.role === ROLE.Admin ||
              authCtx.role === ROLE.Operator ||
              authCtx.role === ROLE.Voter) ? (
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
