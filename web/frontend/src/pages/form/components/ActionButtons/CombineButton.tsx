import { ShieldCheckIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { OngoingAction, Status } from 'types/form';
import ActionButton from './ActionButton';
import { isManager } from './utils';

const CombineButton = ({ status, handleCombine, ongoingAction, formID }) => {
  const { t } = useTranslation();
  const { authorization, isLogged } = useContext(AuthContext);

  return (
    isManager(formID, authorization, isLogged) &&
    status === Status.PubSharesSubmitted && (
      <ActionButton
        handleClick={handleCombine}
        ongoing={ongoingAction === OngoingAction.Combining}
        ongoingText={t('combining')}>
        <>
          <ShieldCheckIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
          {t('combine')}
        </>
      </ActionButton>
    )
  );
};

export default CombineButton;
