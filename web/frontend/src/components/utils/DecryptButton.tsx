import { KeyIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { STATUS } from 'types/election';
import { ROLE } from 'types/userRole';

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
export default DecryptButton;
