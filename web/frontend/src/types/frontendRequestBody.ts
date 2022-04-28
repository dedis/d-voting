import { ACTION } from './election';
import { ROLE } from './userRole';

interface NewElectionBody {
  Configuration: any;
}

interface NewElectionVoteBody {
  Ballot: [];
}

interface EditElectionBody {
  Action: ACTION.Open | ACTION.Close | ACTION.CombineShares | ACTION.Cancel;
}

interface NewUserRole {
  sciper: string;
  role: ROLE.Admin | ROLE.Operator;
}

interface RemoveUserRole {
  sciper: string;
}

export type { NewElectionVoteBody, NewElectionBody, EditElectionBody, NewUserRole, RemoveUserRole };
