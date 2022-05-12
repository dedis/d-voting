import { EyeOffIcon } from '@heroicons/react/outline';
import { IndigoSpinnerIcon } from 'components/utils/SpinnerIcon';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { OngoingAction, Status } from 'types/election';
import { UserRole } from 'types/userRole';

const ShuffleButton = ({ status, handleShuffle, ongoingAction }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === UserRole.Admin || authCtx.role === UserRole.Operator;

  return (
    isAuthorized &&
    status === Status.Closed && (
      <button onClick={handleShuffle}>
        {ongoingAction === OngoingAction.None && (
          <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 border border-gray-300 text-sm rounded-full font-medium text-gray-700 hover:text-indigo-500">
            <EyeOffIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
            {t('shuffle')}
          </div>
        )}
        {ongoingAction === OngoingAction.Shuffling && (
          <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
            <IndigoSpinnerIcon /> {t('shuffling')}
          </div>
        )}
      </button>
    )
  );
};

export default ShuffleButton;
