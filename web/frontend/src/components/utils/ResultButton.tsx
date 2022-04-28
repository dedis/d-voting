import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { STATUS } from 'types/election';

const ResultButton = ({ status, electionID }) => {
  const { t } = useTranslation();
  return (
    status === STATUS.ResultAvailable && (
      <Link to={`/elections/${electionID}/result`}>
        <button>{t('seeResult')}</button>
      </Link>
    )
  );
};
export default ResultButton;
