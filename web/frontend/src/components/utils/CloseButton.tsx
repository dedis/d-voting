import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { STATUS } from 'types/electionInfo';

const CloseButton = ({ status, handleClose }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === 'admin' || authCtx.role === 'operator';

  return (
    isAuthorized && status === STATUS.OPEN && <button onClick={handleClose}>{t('close')}</button>
  );
};
export default CloseButton;
