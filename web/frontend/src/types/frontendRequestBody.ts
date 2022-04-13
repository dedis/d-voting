import { ID } from './configuration';

interface GetElectionBody {
  ElectionID: ID;
  Token: string;
}

interface CreateElectionBody {
  Configuration: any;
}

interface CreateElectionCastVote {
  UserID: string;
  Ballot: [];
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
  GetElectionBody,
  CreateElectionBody,
  GetAllElections,
  LightElectionInfo,
  ElectionInfo,
};
