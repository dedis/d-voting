import { TrashIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';

const SUBJECT_ELECTION = 'election';
const ACTION_CREATE = 'create';

const DeleteButton = ({ handleDelete }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  return (
    authCtx.isAllowed(authCtx, SUBJECT_ELECTION, ACTION_CREATE) && (
      <button onClick={handleDelete}>
        <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700 hover:text-red-500">
          <TrashIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
          {t('delete')}
        </div>
      </button>
    )
  );
};
export default DeleteButton;
