import { ID } from './configuration';

export const enum STATUS {
  // Initial is when the node has been initialized by the dkg service
  Initial = 0,
  // BeingSetup is when a node is currently being setup by the dkg service
  BeingSetup = 7,
  // Setup is when a node has been setup by the dkg service
  Setup = 8,
  // Open is when an election is open, and users can cast ballots
  Open = 1,
  // Closed is when an election is closed, users can no longer cast ballots
  Closed = 2,
  // BeingShuffled is when the ballots are currently being shuffled
  BeingShuffled = 9,
  // ShuffledBallots is when the ballots have been shuffled
  ShuffledBallots = 3,
  // BeingDecrypted is when public keys are currently being shared and combined
  BeingDecrypted = 10,
  // DecryptedBallots is when public keys have been shared and combined
  DecryptedBallots = 4,
  // ResultAvailable is when the ballots have been decrypted
  ResultAvailable = 5,
  // Canceled is when an election has been canceled
  Canceled = 6,
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

interface DownloadedResults {
  ID: ID;
  Title: string;
  Results?: { Candidate: string; Percentage: string }[];
}

export type {
  LightElectionInfo,
  ElectionInfo,
  RankResults,
  Results,
  TextResults,
  SelectResults,
  DownloadedResults,
};
