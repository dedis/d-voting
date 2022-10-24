import { LockOpenIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { OngoingAction, Status } from 'types/form';
import { UserRole } from 'types/userRole';
import ActionButton from './ActionButton';

const OpenButton = ({ status, handleOpen, ongoingAction }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === UserRole.Admin || authCtx.role === UserRole.Operator;

  return (
    isAuthorized &&
    status === Status.Setup && (
      <ActionButton
        handleClick={handleOpen}
        ongoing={ongoingAction === OngoingAction.Opening}
        ongoingText={t('opening')}>
        <>
          <LockOpenIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
          {t('open')}
        </>
      </ActionButton>
    )
  );
};

export default OpenButton;
