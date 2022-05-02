import { EyeOffIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { STATUS } from 'types/election';
import { ROLE } from 'types/userRole';

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
export default ShuffleButton;
