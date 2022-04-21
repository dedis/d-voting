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
  Result: any;
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

interface Results {
  SelectResultIDs: ID[];
  SelectResult: boolean[][];
  RankResultIDs: ID[];
  RankResult: number[][];
  TextResultIDs: ID[];
  TextResult: string[][];
}

type SelectResults = Map<ID, number[][]>;

type RankResults = Map<ID, number[][]>;

type TextResults = Map<ID, string[][]>;

export type { LightElectionInfo, ElectionInfo, RankResults, Results, TextResults, SelectResults };
