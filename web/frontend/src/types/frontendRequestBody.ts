import { ID } from './configuration';
import { Action } from './election';
import { UserRole } from './userRole';

interface NewElectionBody {
  Configuration: any;
}

interface NewElectionVoteBody {
  Ballot: [];
}

interface EditElectionBody {
  Action: Action.Open | Action.Close | Action.CombineShares | Action.Cancel;
}

interface EditDKGActorBody {
  Action: Action.Setup | Action.BeginDecryption;
}

interface NewDKGBody {
  ElectionID: ID;
}

interface NewUserRole {
  sciper: string;
  role: UserRole.Admin | UserRole.Operator;
}

interface RemoveUserRole {
  sciper: string;
}

export type {
  NewElectionVoteBody,
  NewElectionBody,
  EditElectionBody,
  EditDKGActorBody,
  NewDKGBody,
  NewUserRole,
  RemoveUserRole,
};
