import { ID } from './configuration';

interface GetElectionBody {
  ElectionID: ID;
  Token: string;
}

interface CreateElectionBody {
  Format: any;
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

type GetAllElections = LightElectionInfo[];

export type {
  CreateElectionCastVote,
  GetElectionBody,
  CreateElectionBody,
  GetAllElections,
  LightElectionInfo,
};
