import { PencilAltIcon } from '@heroicons/react/outline';
import { AuthContext } from 'index';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { ROUTE_BALLOT_SHOW } from 'Routes';
import { STATUS } from 'types/election';
import { ROLE } from 'types/userRole';

const VoteButton = ({ status, electionID }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const isAuthorized =
    authCtx.role === ROLE.Admin || authCtx.role === ROLE.Operator || authCtx.role === ROLE.Voter;

  return (
    isAuthorized &&
    status === STATUS.Open &&
    authCtx.isLogged && (
      <Link to={ROUTE_BALLOT_SHOW + '/' + electionID}>
        <button>
          <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
            <PencilAltIcon className="-ml-1 mr-2 h-5 w-5 text-gray-700" aria-hidden="true" />
            {t('navBarVote')}
          </div>
        </button>
      </Link>
    )
  );
};
export default VoteButton;
