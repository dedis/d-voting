import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { STATUS } from 'types/electionInfo';
import { Role } from 'types/userRole';

const ShuffleButton = ({ status, isShuffling, handleShuffle }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized = authCtx.role === Role.Admin || authCtx.role === Role.Operator;

  return (
    isAuthorized &&
    status === STATUS.Closed &&
    (isShuffling ? (
      <p className="loading">{t('shuffleOnGoing')}</p>
    ) : (
      <span>
        <button onClick={handleShuffle}>{t('shuffle')}</button>
      </span>
    ))
  );
};
export default ShuffleButton;
