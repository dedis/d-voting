import { AuthContext } from 'index';
import { useContext } from 'react';
import { STATUS } from 'types/electionInfo';

const CloseButton = ({ status, handleClose, t }) => {
  const authCtx = useContext(AuthContext);

  const isAuthorized = authCtx.role === 'admin' || authCtx.role === 'operator';

  return (
    isAuthorized && status === STATUS.OPEN && <button onClick={handleClose}>{t('close')}</button>
  );
};
export default CloseButton;
