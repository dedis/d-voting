import { Configuration } from './configuration';

interface GetElectionBody {
  ElectionID: string;
  Token: string;
}

interface CreateElectionBody {
  Title: string;
  AdminID: string;
  Token: string;
  Format: Configuration;
}

export type { GetElectionBody, CreateElectionBody };
