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
        <button>{t('navBarVote')}</button>
      </Link>
    )
  );
};
export default VoteButton;
