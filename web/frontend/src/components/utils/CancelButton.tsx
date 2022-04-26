import { AuthContext } from 'index';
import { useContext } from 'react';
import { STATUS } from 'types/electionInfo';

const CancelButton = ({ status, handleCancel, t }) => {
  const authCtx = useContext(AuthContext);

  const isAuthorized = authCtx.role === 'admin' || authCtx.role === 'operator';

  return (
    isAuthorized && status === STATUS.OPEN && <button onClick={handleCancel}>{t('cancel')}</button>
  );
};
export default CancelButton;
