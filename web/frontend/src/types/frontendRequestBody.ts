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

interface NewElectionBody {
  Configuration: any;
}

interface NewElectionVoteBody {
  Ballot: [];
}

interface EditElectionBody {
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
  NewElectionVoteBody,
  NewElectionBody,
  GetAllElections,
  LightElectionInfo,
  EditElectionBody,
};

export { STATUS };
