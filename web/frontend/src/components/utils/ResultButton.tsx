import { ChartSquareBarIcon } from '@heroicons/react/outline';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { Status } from 'types/election';

const ResultButton = ({ status, electionID }) => {
  const { t } = useTranslation();
  return (
    status === Status.ResultAvailable && (
      <Link to={`/elections/${electionID}/result`}>
        <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
          <ChartSquareBarIcon className="-ml-1 mr-2 h-5 w-5 text-gray-700" aria-hidden="true" />
          {t('seeResult')}
        </div>
      </Link>
    )
  );
};

export default ResultButton;
