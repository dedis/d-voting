import { CubeTransparentIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { OngoingAction, Status } from 'types/form';
import ActionButton from './ActionButton';
import { isManager } from './utils';

const InitializeButton = ({ status, handleInitialize, ongoingAction, formID }) => {
  const { authorization, isLogged } = useContext(AuthContext);
  const { t } = useTranslation();

  return (
    isManager(formID, authorization, isLogged) &&
    status === Status.Initial && (
      <ActionButton
        handleClick={handleInitialize}
        ongoing={ongoingAction === OngoingAction.Initializing}
        ongoingText={t('initializing')}>
        <>
          <CubeTransparentIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
          {t('initializeNode')}
        </>
      </ActionButton>
    )
  );
};

export default InitializeButton;
