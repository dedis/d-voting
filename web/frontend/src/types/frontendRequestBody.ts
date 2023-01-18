import { ID } from './configuration';
import { Action } from './form';
import { UserRole } from './userRole';

interface NewFormBody {
  Configuration: any;
}

interface NewFormVoteBody {
  Ballot: [];
}

interface EditFormBody {
  Action: Action.Open | Action.Close | Action.CombineShares | Action.Cancel;
}

interface EditDKGActorBody {
  Action: Action.Setup | Action.BeginDecryption;
  Proxy: string;
}

interface NewDKGBody {
  FormID: ID;
  Proxy: string;
}

interface NewUserRole {
  sciper: string;
  role: UserRole.Admin | UserRole.Operator;
}

interface RemoveUserRole {
  sciper: string;
}

interface NewProxyAddress {
  NodeAddr: string;
  Proxy: string;
}

interface UpdateProxyAddress {
  Proxy: string;
  NewNode: string;
}

export type {
  NewFormVoteBody,
  NewFormBody,
  EditFormBody,
  EditDKGActorBody,
  NewDKGBody,
  NewProxyAddress,
  UpdateProxyAddress,
  NewUserRole,
  RemoveUserRole,
};
