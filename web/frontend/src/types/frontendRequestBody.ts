import { Role } from './userRole';

interface NewElectionBody {
  Configuration: any;
}

interface NewElectionVoteBody {
  Ballot: [];
}

interface EditElectionBody {
  Action: 'open' | 'close' | 'combineShares' | 'cancel';
}

interface NewUserRole {
  sciper: string;
  role: Role.Admin | Role.Operator;
}

interface RemoveUserRole {
  sciper: string;
}

export type { NewElectionVoteBody, NewElectionBody, EditElectionBody, NewUserRole, RemoveUserRole };
