import { DocumentAddIcon } from '@heroicons/react/outline';
import { useTranslation } from 'react-i18next';
import { isManager } from './../../../../utils/auth';
import { AuthContext } from 'index';
import { useContext } from 'react';
import IndigoSpinnerIcon from '../IndigoSpinnerIcon';
import { OngoingAction } from 'types/form';

const AddVotersButton = ({ handleAddVoters, formID, ongoingAction }) => {
  const { t } = useTranslation();
  const { authorization, isLogged } = useContext(AuthContext);

  return ongoingAction !== OngoingAction.AddVoters ? (
    isManager(formID, authorization, isLogged) && (
      <button data-testid="addVotersButton" onClick={handleAddVoters}>
        <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700 hover:text-red-500">
          <DocumentAddIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
          {t('addVoters')}
        </div>
      </button>
    )
  ) : (
    <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
      <IndigoSpinnerIcon />
      {t('addVotersLoading')}
    </div>
  );
};
export default AddVotersButton;
