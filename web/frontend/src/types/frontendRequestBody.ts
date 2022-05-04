import { ID } from './configuration';
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

interface EditDKGActorBody {
  Action: ACTION.Setup | ACTION.BeginDecryption;
}

interface NewDKGBody {
  ElectionID: ID;
}

interface NewUserRole {
  sciper: string;
  role: ROLE.Admin | ROLE.Operator;
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
