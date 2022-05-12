import { LockClosedIcon } from '@heroicons/react/outline';
import { IndigoSpinnerIcon } from 'components/utils/SpinnerIcon';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { OngoingAction, Status } from 'types/election';
import { UserRole } from 'types/userRole';

const CloseButton = ({ status, handleClose, ongoingAction }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === UserRole.Admin || authCtx.role === UserRole.Operator;

  return (
    isAuthorized &&
    status === Status.Open && (
      <button onClick={handleClose}>
        {ongoingAction !== OngoingAction.Closing && (
          <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700 hover:text-indigo-500">
            <LockClosedIcon className="-ml-1 mr-2 h-5 w-" aria-hidden="true" />
            {t('close')}
          </div>
        )}
        {ongoingAction === OngoingAction.Closing && (
          <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
            <IndigoSpinnerIcon />
            {t('closing')}
          </div>
        )}
      </button>
    )
  );
};

export default CloseButton;
