import { ID } from './configuration';

export const enum STATUS {
  INITIAL,
  OPEN,
  CLOSED,
  SHUFFLED_BALLOTS,
  DECRYPTED_BALLOTS,
  RESULT_AVAILABLE,
  CANCELED,
}

interface ElectionInfo {
  ElectionID: ID;
  Status: STATUS;
  Pubkey: string;
  Result: [];
  ChunksPerBallot: number;
  BallotSize: number;
  Configuration: any;
}

interface LightElectionInfo {
  ElectionID: ID;
  Title: string;
  Status: STATUS;
  Pubkey: string;
}

interface Result {
  SelectResultIDs: ID[];
  SelectResult: boolean[][];
  RankResultIDs: ID[];
  RankResult: number[][];
  TextResultIDs: ID[];
  TextResult: string[][];
}

export type { LightElectionInfo, ElectionInfo, Result };
