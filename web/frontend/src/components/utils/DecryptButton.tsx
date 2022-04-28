import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { STATUS } from 'types/electionInfo';

const DecryptButton = ({ status, isDecrypting, handleDecrypt }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === 'admin' || authCtx.role === 'operator';

  return (
    isAuthorized &&
    status === STATUS.SHUFFLED_BALLOTS &&
    (isDecrypting ? (
      <p className="loading">{t('decryptOnGoing')}</p>
    ) : (
      <span>
        <button onClick={handleDecrypt}>{t('decrypt')}</button>
      </span>
    ))
  );
};
export default DecryptButton;
