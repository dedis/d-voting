import { ID } from './configuration';

interface ElectionInfo {
  ElectionID: ID;
  Status: number;
  Pubkey: string;
  Result: [];
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

// TODO change to Map, requires to unmarshal the object in useElection
interface Result {
  SelectResultIDs: ID[];
  SelectResult: [boolean[]];
  RankResultIDs: ID[];
  RankResult: [number[]];
  TextResultIDs: ID[];
  TextResult: [string[]];
}

export type { LightElectionInfo, ElectionInfo, Result };
