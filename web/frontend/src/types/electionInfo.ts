import { ID } from './configuration';

interface ElectionInfo {
  ElectionID: ID;
  Status: number;
  Pubkey: string;
  Result: any;
  ChunksPerBallot: number;
  BallotSize: number;
  Configuration: any;
}

interface LightElectionInfo {
  ElectionID: ID;
  Title: string;
  Status: number;
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
