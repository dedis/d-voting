import { PencilAltIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { ROUTE_BALLOT_SHOW } from 'Routes';
import { Status } from 'types/form';

const VoteButton = ({ status, formID }) => {
  const { isLogged } = useContext(AuthContext);
  const { t } = useTranslation();

  return (
    status === Status.Open &&
    isLogged && (
      <Link to={ROUTE_BALLOT_SHOW + '/' + formID}>
        <button>
          <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700 hover:text-[#ff0000]">
            <PencilAltIcon className="sm:-ml-1 sm:mr-2 h-5 w-5" aria-hidden="true" />
            <div className="hidden sm:block">{t('vote')}</div>
          </div>
        </button>
      </Link>
    )
  );
};

export default VoteButton;
