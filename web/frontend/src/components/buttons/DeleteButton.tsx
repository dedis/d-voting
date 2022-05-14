import { TrashIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { UserRole } from 'types/userRole';

const DeleteButton = ({ status, handleDelete }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === UserRole.Admin || authCtx.role === UserRole.Operator;

  return isAuthorized ? (
    <button onClick={handleDelete}>
      <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
        <TrashIcon className="-ml-1 mr-2 h-5 w-5 text-gray-700" aria-hidden="true" />
        {t('delete')}
      </div>
    </button>
  ) : (
    <></>
  );
};
export default DeleteButton;
