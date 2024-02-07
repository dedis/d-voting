import { EyeOffIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { OngoingAction, Status } from 'types/form';
import ActionButton from './ActionButton';
import { isManager } from './utils';

const ShuffleButton = ({ status, handleShuffle, ongoingAction, formID }) => {
  const { t } = useTranslation();
  const { authorization, isLogged } = useContext(AuthContext);

  return (
    isManager(formID, authorization, isLogged) &&
    status === Status.Closed && (
      <ActionButton
        handleClick={handleShuffle}
        ongoing={ongoingAction === OngoingAction.Shuffling}
        ongoingText={t('shuffling')}>
        <>
          <EyeOffIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
          {t('shuffle')}
        </>
      </ActionButton>
    )
  );
};

export default ShuffleButton;
