import { KeyIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { OngoingAction, Status } from 'types/election';
import { UserRole } from 'types/userRole';
import ActionButton from './ActionButton';

const DecryptButton = ({ status, handleDecrypt, ongoingAction }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === UserRole.Admin || authCtx.role === UserRole.Operator;

  return (
    isAuthorized &&
    status === Status.ShuffledBallots && (
      <ActionButton
        handleClick={handleDecrypt}
        ongoing={ongoingAction === OngoingAction.Decrypting}
        ongoingText={t('decrypting')}>
        <>
          <KeyIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
          {t('decrypt')}
        </>
      </ActionButton>
    )
  );
};

export default DecryptButton;
