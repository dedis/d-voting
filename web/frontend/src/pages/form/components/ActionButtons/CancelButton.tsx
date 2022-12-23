import { XIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { OngoingAction, Status } from 'types/form';
import ActionButton from './ActionButton';
function hasAuthorization(authCtx, subject: string, action: string): boolean {
  return (
    authCtx.authorization.has(subject) && authCtx.authorization.get(subject).indexOf(action) !== -1
  );
}
const CancelButton = ({ status, handleCancel, ongoingAction }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  //const isAuthorized = authCtx.role === 'admin' || authCtx.role === 'operator';

  return (
    hasAuthorization(authCtx, 'election', 'create') &&
    status === Status.Open && (
      <ActionButton
        handleClick={handleCancel}
        ongoing={ongoingAction === OngoingAction.Canceling}
        ongoingText={t('canceling')}>
        <>
          <XIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
          {t('cancel')}
        </>
      </ActionButton>
    )
  );
};

export default CancelButton;
