import {
  ChartSquareBarIcon,
  EyeOffIcon,
  KeyIcon,
  LockClosedIcon,
  PencilAltIcon,
  ShieldCheckIcon,
  XIcon,
} from '@heroicons/react/outline';
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
    status === STATUS.Initialized && <button onClick={handleSetup}>{t('statusSetup')}</button>
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

const VoteButton = ({ status, electionID }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized =
    authCtx.role === ROLE.Admin || authCtx.role === ROLE.Operator || authCtx.role === ROLE.Voter;

  return (
    isAuthorized &&
    status === STATUS.Open &&
    authCtx.isLogged && (
      <Link to={ROUTE_BALLOT_SHOW + '/' + electionID}>
        <button>
          <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
            <PencilAltIcon className="-ml-1 mr-2 h-5 w-5 text-gray-700" aria-hidden="true" />
            {t('vote')}
          </div>
        </button>
      </Link>
    )
  );
};

const CloseButton = ({ status, handleClose }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === ROLE.Admin || authCtx.role === ROLE.Operator;

  return (
    isAuthorized &&
    status === STATUS.Open && (
      <button onClick={handleClose}>
        <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
          <LockClosedIcon className="-ml-1 mr-2 h-5 w-5 text-gray-700" aria-hidden="true" />
          {t('close')}
        </div>
      </button>
    )
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
        <button onClick={handleShuffle}>
          <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
            <EyeOffIcon className="-ml-1 mr-2 h-5 w-5 text-gray-700" aria-hidden="true" />
            {t('shuffle')}
          </div>
        </button>
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
        <button onClick={handleDecrypt}>
          <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
            <KeyIcon className="-ml-1 mr-2 h-5 w-5 text-gray-700" aria-hidden="true" />
            {t('decrypt')}
          </div>
        </button>
      </span>
    ))
  );
};

const CombineButton = ({ status, handleCombine }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === ROLE.Admin || authCtx.role === ROLE.Operator;

  return (
    isAuthorized &&
    status === STATUS.PubSharesSubmitted && (
      <span>
        <button onClick={handleCombine}>
          <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
            <ShieldCheckIcon className="-ml-1 mr-2 h-5 w-5 text-gray-700" aria-hidden="true" />
            {t('combine')}
          </div>
        </button>
      </span>
    )
  );
};

const ResultButton = ({ status, electionID }) => {
  const { t } = useTranslation();
  return (
    status === STATUS.ResultAvailable && (
      <Link to={`/elections/${electionID}/result`}>
        <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
          <ChartSquareBarIcon className="-ml-1 mr-2 h-5 w-5 text-gray-700" aria-hidden="true" />
          {t('seeResult')}
        </div>
      </Link>
    )
  );
};

const CancelButton = ({ status, handleCancel }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === 'admin' || authCtx.role === 'operator';

  return (
    isAuthorized &&
    status === STATUS.Open && (
      <button onClick={handleCancel}>
        <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
          <XIcon className="-ml-1 mr-2 h-5 w-5 text-gray-700" aria-hidden="true" />
          {t('cancel')}
        </div>
      </button>
    )
  );
};

export {
  InitializeButton,
  SetupButton,
  OpenButton,
  VoteButton,
  CloseButton,
  ShuffleButton,
  DecryptButton,
  CombineButton,
  ResultButton,
  CancelButton,
};
