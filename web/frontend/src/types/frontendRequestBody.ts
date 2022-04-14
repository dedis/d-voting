interface CreateElectionBody {
  Configuration: any;
}

interface CreateElectionCastVote {
  Ballot: [];
}

interface ElectionActionsBody {
  Action: 'open' | 'close' | 'combineShares' | 'cancel';
}

export type { CreateElectionCastVote, CreateElectionBody, ElectionActionsBody };
