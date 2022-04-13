import { ID } from './configuration';

interface CreateElectionBody {
  UserID: string;
  Configuration: any;
}

interface CreateElectionCastVote {
  UserID: string;
  Ballot: [];
}

interface ElectionActionsBody {
  Action: 'open' | 'close' | 'combineShares' | 'cancel';
}

interface LightElectionInfo {
  ElectionID: ID;
  Title: string;
  Status: number;
  Pubkey: string;
}

type GetAllElections = LightElectionInfo[];

export type {
  CreateElectionCastVote,
  CreateElectionBody,
  GetAllElections,
  LightElectionInfo,
  ElectionActionsBody,
};
