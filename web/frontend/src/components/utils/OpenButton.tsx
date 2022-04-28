import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { STATUS } from 'types/electionInfo';
import { Role } from 'types/userRole';

const OpenButton = ({ status, handleOpen }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === Role.Admin || authCtx.role === Role.Operator;

  return (
    isAuthorized && status === STATUS.Initial && <button onClick={handleOpen}>{t('open')}</button>
  );
};

export default OpenButton;
