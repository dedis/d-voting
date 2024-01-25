import { DocumentAddIcon } from '@heroicons/react/outline';
import { useTranslation } from 'react-i18next';
import { isManager } from './utils';
import { AuthContext } from 'index';
import { useContext } from 'react';

const AddVotersButton = ({ handleAddVoters, formID }) => {
  const { t } = useTranslation();
  const { authorization, isLogged } = useContext(AuthContext);

  return (
    isManager(formID, authorization, isLogged) && (
      <button data-testid="addVotersButton" onClick={handleAddVoters}>
        <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700 hover:text-red-500">
          <DocumentAddIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
          {t('addVoters')}
        </div>
      </button>
    )
  );
};
export default AddVotersButton;
