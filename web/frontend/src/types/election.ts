import { ID } from './configuration';

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

export type { LightElectionInfo, ElectionInfo };
