import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { STATUS } from 'types/election';
import { ROLE } from 'types/userRole';

const CloseButton = ({ status, handleClose }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === ROLE.Admin || authCtx.role === ROLE.Operator;

  return (
    isAuthorized && status === STATUS.Open && <button onClick={handleClose}>{t('close')}</button>
  );
};
export default CloseButton;
