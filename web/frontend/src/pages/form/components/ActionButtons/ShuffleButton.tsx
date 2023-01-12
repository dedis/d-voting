import { EyeOffIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { OngoingAction, Status } from 'types/form';
import ActionButton from './ActionButton';

const SUBJECT_ELECTION = 'election';
const ACTION_CREATE = 'create';
const ShuffleButton = ({ status, handleShuffle, ongoingAction }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  return (
    authCtx.isAllowed(authCtx, SUBJECT_ELECTION, ACTION_CREATE) &&
    status === Status.Closed && (
      <ActionButton
        handleClick={handleShuffle}
        ongoing={ongoingAction === OngoingAction.Shuffling}
        ongoingText={t('shuffling')}>
        <>
          <EyeOffIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
          {t('shuffle')}
        </>
      </ActionButton>
    )
  );
};

export default ShuffleButton;
