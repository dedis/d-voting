import { PencilAltIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { ROUTE_BALLOT_SHOW } from 'Routes';
import { Status } from 'types/election';
import { UserRole } from 'types/userRole';

const VoteButton = ({ status, electionID }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized =
    authCtx.role === UserRole.Admin ||
    authCtx.role === UserRole.Operator ||
    authCtx.role === UserRole.Voter;

  return (
    isAuthorized &&
    status === Status.Open &&
    authCtx.isLogged && (
      <Link to={ROUTE_BALLOT_SHOW + '/' + electionID}>
        <button>
          <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 border border-gray-300 text-sm rounded-full font-medium text-gray-700 hover:text-indigo-500">
            <PencilAltIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
            {t('vote')}
          </div>
        </button>
      </Link>
    )
  );
};

export default VoteButton;
