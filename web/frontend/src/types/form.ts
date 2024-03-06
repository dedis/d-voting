import { ID, Title } from './configuration';

export const enum Status {
  // Initial is when the form has just been created
  Initial = 0,
  // InitializedNode is when the nodes has been initialized by the dkg service
  Initialized = 7,
  // Setup is when a node has been setup by the dkg service
  Setup = 8,
  // Open is when a form is open, and users can cast ballots
  Open = 1,
  // Closed is when a form is closed, users can no longer cast ballots
  Closed = 2,
  // ShuffledBallots is when the ballots have been shuffled
  ShuffledBallots = 3,
  // DecryptedBallots is when public keys have been shared and combined
  PubSharesSubmitted = 4,
  // ResultAvailable is when the ballots have been decrypted
  ResultAvailable = 5,
  // Canceled is when a form has been canceled
  Canceled = 6,
}

export const enum Action {
  Initialize = 'initialize',
  Setup = 'setup',
  Open = 'open',
  Close = 'close',
  Shuffle = 'shuffle',
  BeginDecryption = 'computePubshares',
  CombineShares = 'combineShares',
  Cancel = 'cancel',
}

export const enum OngoingAction {
  None,
  Initializing,
  SettingUp,
  Opening,
  Closing,
  Shuffling,
  Decrypting,
  Combining,
  Canceling,
}

interface FormInfo {
  FormID: ID;
  Status: Status;
  Pubkey: string;
  Result: any;
  Roster: string[];
  ChunksPerBallot: number;
  BallotSize: number;
  Configuration: any;
  Voters: string[];
}

interface LightFormInfo {
  FormID: ID;
  Title: Title;
  Status: Status;
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
  Title: string;
  URL: string;
  Results?: {
    Candidate: string;
    Percent?: string;
    TotalCount?: number;
    NumberOfBallots?: number;
  }[];
}
interface BallotResults {
  BallotNumber: number;
  Results: DownloadedResults[];
}

export type {
  LightFormInfo,
  FormInfo,
  RankResults,
  Results,
  TextResults,
  SelectResults,
  DownloadedResults,
  BallotResults,
};
