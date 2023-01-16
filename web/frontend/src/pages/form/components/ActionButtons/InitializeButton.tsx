import { CubeTransparentIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { OngoingAction, Status } from 'types/form';
import ActionButton from './ActionButton';
const SUBJECT_ELECTION = 'election';
const ACTION_CREATE = 'create';
const InitializeButton = ({ status, handleInitialize, ongoingAction }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  return (
    authCtx.isAllowed(SUBJECT_ELECTION, ACTION_CREATE) &&
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
