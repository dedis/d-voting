import { AuthContext } from 'index';
import { useContext } from 'react';
import { STATUS } from 'types/electionInfo';

const DecryptButton = ({ status, isDecrypting, handleDecrypt, t }) => {
  const authCtx = useContext(AuthContext);

  const isAuthorized = authCtx.role === 'admin' || authCtx.role === 'operator';

  return isAuthorized && status === STATUS.SHUFFLED_BALLOTS && isDecrypting ? (
    <p className="loading">{t('decryptOnGoing')}</p>
  ) : (
    <span>
      <button onClick={handleDecrypt}>{t('decrypt')}</button>
    </span>
  );
};
export default DecryptButton;
