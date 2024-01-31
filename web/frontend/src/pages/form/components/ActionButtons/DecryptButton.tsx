import { KeyIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { OngoingAction, Status } from 'types/form';
import ActionButton from './ActionButton';
import { isManager } from './utils';

const DecryptButton = ({ status, handleDecrypt, ongoingAction, formID }) => {
  const { authorization, isLogged } = useContext(AuthContext);
  const { t } = useTranslation();

  return (
    isManager(formID, authorization, isLogged) &&
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
