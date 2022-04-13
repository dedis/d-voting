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

interface ElectionInfo {
  ElectionID: ID;
  Status: number;
  Pubkey: string;
  Result: [];
  ChunksPerBallot: number;
  BallotSize: number;
  Configuration: any;
}

type GetAllElections = LightElectionInfo[];

export type {
  CreateElectionCastVote,
  CreateElectionBody,
  GetAllElections,
  LightElectionInfo,
  ElectionInfo,
  ElectionActionsBody,
};
