import { ID } from './configuration';

interface CreateElectionBody {
  Configuration: any;
}

interface CreateElectionCastVote {
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
