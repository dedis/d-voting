import { XIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { OngoingAction, Status } from 'types/form';
import ActionButton from './ActionButton';
const SUBJECT_ELECTION = 'election';
const ACTION_CREATE = 'create';
const CancelButton = ({ status, handleCancel, ongoingAction }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  return (
    authCtx.isAllowed(authCtx, SUBJECT_ELECTION, ACTION_CREATE) &&
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
