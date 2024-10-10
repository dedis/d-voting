import { CogIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { OngoingAction, Status } from 'types/form';
import ActionButton from './ActionButton';
import { isManager } from './../../../../utils/auth';

const SetupButton = ({ status, handleSetup, ongoingAction, formID }) => {
  const { t } = useTranslation();
  const { authorization, isLogged } = useContext(AuthContext);

  return (
    isManager(formID, authorization, isLogged) &&
    status === Status.Initialized && (
      <ActionButton
        handleClick={handleSetup}
        ongoing={ongoingAction === OngoingAction.SettingUp}
        ongoingText={t('settingUp')}>
        <>
          <CogIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
          {t('statusSetup')}
        </>
      </ActionButton>
    )
  );
};

export default SetupButton;
