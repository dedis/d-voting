import { CubeTransparentIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { OngoingAction, Status } from 'types/election';
import { UserRole } from 'types/userRole';
import ActionButton from './ActionButton';

const InitializeButton = ({ status, handleInitialize, ongoingAction }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === UserRole.Admin || authCtx.role === UserRole.Operator;

  return (
    isAuthorized &&
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
