import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { STATUS } from 'types/electionInfo';

const ResultButton = ({ status, electionID }) => {
  const { t } = useTranslation();
  return (
    status === STATUS.RESULT_AVAILABLE && (
      <Link to={`/elections/${electionID}`}>
        <button>{t('seeResult')}</button>
      </Link>
    )
  );
};
export default ResultButton;
