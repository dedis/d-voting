import { XIcon } from '@heroicons/react/outline';
import { IndigoSpinnerIcon } from 'components/utils/SpinnerIcon';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { OngoingAction, Status } from 'types/election';

const CancelButton = ({ status, handleCancel, ongoingAction }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === 'admin' || authCtx.role === 'operator';

  return (
    isAuthorized &&
    status === Status.Open && (
      <button onClick={handleCancel}>
        {ongoingAction !== OngoingAction.Canceling && (
          <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700 hover:text-indigo-500">
            <XIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
            {t('cancel')}
          </div>
        )}
        {ongoingAction === OngoingAction.Canceling && (
          <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
            <IndigoSpinnerIcon />
            {t('canceling')}
          </div>
        )}
      </button>
    )
  );
};

export default CancelButton;
