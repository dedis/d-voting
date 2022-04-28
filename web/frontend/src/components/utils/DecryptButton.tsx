import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { STATUS } from 'types/electionInfo';
import { Role } from 'types/userRole';

const DecryptButton = ({ status, isDecrypting, handleDecrypt }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === Role.Admin || authCtx.role === Role.Operator;

  return (
    isAuthorized &&
    status === STATUS.ShuffledBallots &&
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
