import { LockClosedIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { OngoingAction, Status } from 'types/form';
import ActionButton from './ActionButton';

const SUBJECT_ELECTION = 'election';
const ACTION_CREATE = 'create';
const CloseButton = ({ status, handleClose, ongoingAction }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  return (
    authCtx.isAllowed(authCtx, SUBJECT_ELECTION, ACTION_CREATE) &&
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
