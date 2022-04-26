import { AuthContext } from 'index';
import { useContext } from 'react';
import { STATUS } from 'types/electionInfo';

const OpenButton = ({ status, handleOpen, t }) => {
  const authCtx = useContext(AuthContext);

  const isAuthorized = authCtx.role === 'admin' || authCtx.role === 'operator';

  return (
    isAuthorized && status === STATUS.INITIAL && <button onClick={handleOpen}>{t('open')}</button>
  );
};

export default OpenButton;
