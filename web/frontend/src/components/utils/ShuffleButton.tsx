import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { STATUS } from 'types/election';
import { ROLE } from 'types/userRole';

const ShuffleButton = ({ status, isShuffling, handleShuffle }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === ROLE.Admin || authCtx.role === ROLE.Operator;

  return (
    isAuthorized &&
    status === STATUS.Closed &&
    (isShuffling ? (
      <p className="loading">{t('statusOnGoingShuffle')}</p>
    ) : (
      <span>
        <button onClick={handleShuffle}>{t('shuffle')}</button>
      </span>
    ))
  );
};
export default ShuffleButton;
