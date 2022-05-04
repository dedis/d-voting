import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { ROUTE_BALLOT_SHOW } from 'Routes';
import { STATUS } from 'types/election';
import { ROLE } from 'types/userRole';

const InitializeButton = ({ status, handleInitialize }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === ROLE.Admin || authCtx.role === ROLE.Operator;

  return (
    isAuthorized &&
    status === STATUS.Initial && <button onClick={handleInitialize}>{t('initNodes')}</button>
  );
};

const SetupButton = ({ status, handleSetup }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === ROLE.Admin || authCtx.role === ROLE.Operator;

  return (
    isAuthorized &&
    status === STATUS.InitializedNodes && <button onClick={handleSetup}>{t('statusSetup')}</button>
  );
};

const OpenButton = ({ status, handleOpen }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === ROLE.Admin || authCtx.role === ROLE.Operator;

  return (
    isAuthorized && status === STATUS.Setup && <button onClick={handleOpen}>{t('open')}</button>
  );
};

const CastVoteButton = ({ status, electionID }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized =
    authCtx.role === ROLE.Admin || authCtx.role === ROLE.Operator || authCtx.role === ROLE.Voter;

  return (
    isAuthorized &&
    status === STATUS.Open && (
      <Link to={ROUTE_BALLOT_SHOW + '/' + electionID}>{t('navBarVote')}</Link>
    )
  );
};

const CloseButton = ({ status, handleClose }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === ROLE.Admin || authCtx.role === ROLE.Operator;

  return (
    isAuthorized && status === STATUS.Open && <button onClick={handleClose}>{t('close')}</button>
  );
};

const ShuffleButton = ({ status, isShuffling, handleShuffle }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === ROLE.Admin || authCtx.role === ROLE.Operator;

  return (
    isAuthorized &&
    status === STATUS.Closed &&
    (isShuffling ? (
      <p className="loading">{t('statusOnGoingShuffle')}</p>
    ) : (
      <span>
        <button onClick={handleShuffle}>{t('shuffle')}</button>
      </span>
    ))
  );
};

const DecryptButton = ({ status, isDecrypting, handleDecrypt }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === ROLE.Admin || authCtx.role === ROLE.Operator;

  return (
    isAuthorized &&
    status === STATUS.ShuffledBallots &&
    (isDecrypting ? (
      <p className="loading">{t('statusOnGoingDecryption')}</p>
    ) : (
      <span>
        <button onClick={handleDecrypt}>{t('decrypt')}</button>
      </span>
    ))
  );
};

const ResultButton = ({ status, electionID }) => {
  const { t } = useTranslation();
  return (
    status === STATUS.ResultAvailable && (
      <Link to={`/elections/${electionID}/result`}>
        <button>{t('seeResult')}</button>
      </Link>
    )
  );
};

const CancelButton = ({ status, handleCancel }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === 'admin' || authCtx.role === 'operator';

  return (
    isAuthorized && status === STATUS.Open && <button onClick={handleCancel}>{t('cancel')}</button>
  );
};

export {
  InitializeButton,
  SetupButton,
  OpenButton,
  CastVoteButton,
  CloseButton,
  ShuffleButton,
  DecryptButton,
  ResultButton,
  CancelButton,
};
