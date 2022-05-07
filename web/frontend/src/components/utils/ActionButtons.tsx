import {
  ChartSquareBarIcon,
  CogIcon,
  CubeTransparentIcon,
  EyeOffIcon,
  KeyIcon,
  LockClosedIcon,
  LockOpenIcon,
  MinusIcon,
  PencilAltIcon,
  ShieldCheckIcon,
  XIcon,
} from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { ROUTE_BALLOT_SHOW } from 'Routes';
import { OngoingAction, Status } from 'types/election';
import { UserRole } from 'types/userRole';
import { IndigoSpinnerIcon } from './SpinnerIcon';

const InitializeButton = ({ status, handleInitialize, ongoingAction }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === UserRole.Admin || authCtx.role === UserRole.Operator;

  return (
    isAuthorized &&
    status === Status.Initial && (
      <button onClick={handleInitialize}>
        <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
          {ongoingAction === OngoingAction.None && (
            <>
              <CubeTransparentIcon
                className="-ml-1 mr-2 h-5 w-5 text-gray-700"
                aria-hidden="true"
              />
              {t('initializeNode')}
            </>
          )}
          {ongoingAction === OngoingAction.Initializing && (
            <>
              <IndigoSpinnerIcon /> {t('initializing')}
            </>
          )}
        </div>
      </button>
    )
  );
};

const SetupButton = ({ status, handleSetup, ongoingAction }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === UserRole.Admin || authCtx.role === UserRole.Operator;

  return (
    isAuthorized &&
    status === Status.Initialized && (
      <button onClick={handleSetup}>
        <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
          {ongoingAction === OngoingAction.None && (
            <>
              <CogIcon className="-ml-1 mr-2 h-5 w-5 text-gray-700" aria-hidden="true" />
              {t('statusSetup')}
            </>
          )}
          {ongoingAction === OngoingAction.SettingUp && (
            <>
              <IndigoSpinnerIcon /> {t('settingUp')}
            </>
          )}
        </div>
      </button>
    )
  );
};

const OpenButton = ({ status, handleOpen, ongoingAction }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === UserRole.Admin || authCtx.role === UserRole.Operator;

  return (
    isAuthorized &&
    status === Status.Setup && (
      <button onClick={handleOpen}>
        <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
          {ongoingAction === OngoingAction.None && (
            <>
              <LockOpenIcon className="-ml-1 mr-2 h-5 w-5 text-gray-700" aria-hidden="true" />
              {t('open')}
            </>
          )}
          {ongoingAction === OngoingAction.Opening && (
            <>
              <IndigoSpinnerIcon />
              {t('opening')}
            </>
          )}
        </div>
      </button>
    )
  );
};

const VoteButton = ({ status, electionID }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized =
    authCtx.role === UserRole.Admin ||
    authCtx.role === UserRole.Operator ||
    authCtx.role === UserRole.Voter;

  return (
    isAuthorized &&
    status === Status.Open &&
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

const CloseButton = ({ status, handleClose, ongoingAction }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === UserRole.Admin || authCtx.role === UserRole.Operator;

  return (
    isAuthorized &&
    status === Status.Open && (
      <button onClick={handleClose}>
        <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
          {ongoingAction !== OngoingAction.Closing && (
            <>
              <LockClosedIcon className="-ml-1 mr-2 h-5 w-5 text-gray-700" aria-hidden="true" />
              {t('close')}
            </>
          )}
          {ongoingAction === OngoingAction.Closing && (
            <>
              <IndigoSpinnerIcon />
              {t('closing')}
            </>
          )}
        </div>
      </button>
    )
  );
};

const ShuffleButton = ({ status, handleShuffle, ongoingAction }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === UserRole.Admin || authCtx.role === UserRole.Operator;

  return (
    isAuthorized &&
    status === Status.Closed && (
      <button onClick={handleShuffle}>
        <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
          {ongoingAction === OngoingAction.None && (
            <>
              <EyeOffIcon className="-ml-1 mr-2 h-5 w-5 text-gray-700" aria-hidden="true" />
              {t('shuffle')}
            </>
          )}
          {ongoingAction === OngoingAction.Shuffling && (
            <>
              <IndigoSpinnerIcon /> {t('shuffling')}
            </>
          )}
        </div>
      </button>
    )
  );
};

const DecryptButton = ({ status, handleDecrypt, ongoingAction }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === UserRole.Admin || authCtx.role === UserRole.Operator;

  return (
    isAuthorized &&
    status === Status.ShuffledBallots && (
      <button onClick={handleDecrypt}>
        <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
          {ongoingAction === OngoingAction.None && (
            <>
              <KeyIcon className="-ml-1 mr-2 h-5 w-5 text-gray-700" aria-hidden="true" />
              {t('decrypt')}
            </>
          )}
          {ongoingAction === OngoingAction.Decrypting && (
            <>
              <IndigoSpinnerIcon />
              {t('decrypting')}
            </>
          )}
        </div>
      </button>
    )
  );
};

const CombineButton = ({ status, handleCombine, ongoingAction }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === UserRole.Admin || authCtx.role === UserRole.Operator;

  return (
    isAuthorized &&
    status === Status.PubSharesSubmitted && (
      <span>
        <button onClick={handleCombine}>
          <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
            {ongoingAction === OngoingAction.None && (
              <>
                <ShieldCheckIcon className="-ml-1 mr-2 h-5 w-5 text-gray-700" aria-hidden="true" />
                {t('combine')}
              </>
            )}
            {ongoingAction === OngoingAction.Combining && (
              <>
                <IndigoSpinnerIcon />
                {t('combining')}
              </>
            )}
          </div>
        </button>
      </span>
    )
  );
};

const ResultButton = ({ status, electionID }) => {
  const { t } = useTranslation();
  return (
    status === Status.ResultAvailable && (
      <Link to={`/elections/${electionID}/result`}>
        <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
          <ChartSquareBarIcon className="-ml-1 mr-2 h-5 w-5 text-gray-700" aria-hidden="true" />
          {t('seeResult')}
        </div>
      </Link>
    )
  );
};

const CancelButton = ({ status, handleCancel, ongoingAction }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === 'admin' || authCtx.role === 'operator';

  return (
    isAuthorized &&
    status === Status.Open && (
      <button onClick={handleCancel}>
        <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
          {ongoingAction !== OngoingAction.Canceling && (
            <>
              <XIcon className="-ml-1 mr-2 h-5 w-5 text-gray-700" aria-hidden="true" />
              {t('cancel')}
            </>
          )}
          {ongoingAction === OngoingAction.Canceling && (
            <>
              <IndigoSpinnerIcon />
              {t('canceling')}
            </>
          )}
        </div>
      </button>
    )
  );
};

const NoActionAvailable = () => {
  const { t } = useTranslation();

  return (
    <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
      <MinusIcon className="-ml-1 mr-2 h-5 w-5 text-gray-700" aria-hidden="true" />
      {t('noActionAvailable')}
    </div>
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
  NoActionAvailable,
};
