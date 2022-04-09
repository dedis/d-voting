interface GetElectionBody {
  ElectionID: string;
  Token: string;
}

interface CreateElectionBody {
  Format: any;
}

interface GetAllElections {}

export type { GetElectionBody, CreateElectionBody };
