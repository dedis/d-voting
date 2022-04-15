import { ID } from './configuration';

const enum STATUS {
  INITIAL,
  OPEN,
  CLOSED,
  SHUFFLED_BALLOTS,
  DECRYPTED_BALLOTS,
  RESULT_AVAILABLE,
  CANCELED,
}

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
  Status: STATUS;
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

export { STATUS };
