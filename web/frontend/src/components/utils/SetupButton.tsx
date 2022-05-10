import { CogIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { OngoingAction, Status } from 'types/election';
import { UserRole } from 'types/userRole';
import { IndigoSpinnerIcon } from './SpinnerIcon';

const SetupButton = ({ status, handleSetup, ongoingAction }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === UserRole.Admin || authCtx.role === UserRole.Operator;

  return (
    isAuthorized &&
    status === Status.Initialized && (
      <button onClick={handleSetup}>
        {ongoingAction === OngoingAction.None && (
          <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700 hover:text-indigo-500">
            <CogIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
            {t('statusSetup')}
          </div>
        )}
        {ongoingAction === OngoingAction.SettingUp && (
          <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
            <IndigoSpinnerIcon /> {t('settingUp')}
          </div>
        )}
      </button>
    )
  );
};

export default SetupButton;
