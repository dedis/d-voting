interface NewElectionBody {
  Configuration: any;
}

interface NewElectionVoteBody {
  Ballot: [];
}

interface EditElectionBody {
  Action: 'open' | 'close' | 'combineShares' | 'cancel';
}

export type { NewElectionVoteBody, NewElectionBody, EditElectionBody };
