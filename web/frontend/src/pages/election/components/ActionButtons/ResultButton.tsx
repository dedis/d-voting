import { ChartSquareBarIcon } from '@heroicons/react/outline';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { Status } from 'types/election';

const ResultButton = ({ status, electionID }) => {
  const { t } = useTranslation();
  return (
    status === Status.ResultAvailable && (
      <Link to={`/elections/${electionID}/result`}>
        <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700 hover:text-indigo-500">
          <ChartSquareBarIcon className="sm:-ml-1 sm:mr-2 h-5 w-5" aria-hidden="true" />
          <div className="hidden sm:block">{t('seeResult')}</div>
        </div>
      </Link>
    )
  );
};

export default ResultButton;
