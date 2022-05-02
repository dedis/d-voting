import { XIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { STATUS } from 'types/election';

const CancelButton = ({ status, handleCancel }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === 'admin' || authCtx.role === 'operator';

  return (
    isAuthorized &&
    status === STATUS.Open && (
      <button onClick={handleCancel}>
        <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
          <XIcon className="-ml-1 mr-2 h-5 w-5 text-gray-700" aria-hidden="true" />
          {t('cancel')}
        </div>
      </button>
    )
  );
};
export default CancelButton;
