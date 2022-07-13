import { LockClosedIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { OngoingAction, Status } from 'types/election';
import { UserRole } from 'types/userRole';
import ActionButton from './ActionButton';

const CloseButton = ({ status, handleClose, ongoingAction }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === UserRole.Admin || authCtx.role === UserRole.Operator;

  return (
    isAuthorized &&
    status === Status.Open && (
      <ActionButton
        handleClick={handleClose}
        ongoing={ongoingAction === OngoingAction.Closing}
        ongoingText={t('closing')}>
        <>
          <LockClosedIcon className="-ml-1 mr-2 h-5 w-" aria-hidden="true" />
          {t('close')}
        </>
      </ActionButton>
    )
  );
};

export default CloseButton;
