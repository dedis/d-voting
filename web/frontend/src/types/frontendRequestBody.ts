interface GetElectionBody {
  ElectionID: string;
  Token: string;
}

interface CreateElectionBody {
  Format: any;
}

export type { GetElectionBody, CreateElectionBody };
