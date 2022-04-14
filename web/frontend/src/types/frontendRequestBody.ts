import { ID } from './configuration';

const INITIAL_STATUS = 0;
const OPEN_STATUS = 1;
const CLOSED_STATUS = 2;
const SHUFFLED_BALLOTS_STATUS = 3;
const DECRYPTED_BALLOTS_STATUS = 4;
const RESULT_AVAILABLE_STATUS = 5;
const CANCELED_STATUS = 6;

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
  Status:
    | typeof INITIAL_STATUS
    | typeof OPEN_STATUS
    | typeof CLOSED_STATUS
    | typeof SHUFFLED_BALLOTS_STATUS
    | typeof DECRYPTED_BALLOTS_STATUS
    | typeof RESULT_AVAILABLE_STATUS
    | typeof CANCELED_STATUS;
  Pubkey: string;
}

type GetAllElections = LightElectionInfo[];

export type {
  CreateElectionCastVote,
  CreateElectionBody,
  GetAllElections,
  LightElectionInfo,
  ElectionActionsBody,
};

export {
  INITIAL_STATUS,
  OPEN_STATUS,
  CLOSED_STATUS,
  SHUFFLED_BALLOTS_STATUS,
  DECRYPTED_BALLOTS_STATUS,
  RESULT_AVAILABLE_STATUS,
  CANCELED_STATUS,
};
