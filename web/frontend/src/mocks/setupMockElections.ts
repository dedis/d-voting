import ShortUniqueId from 'short-unique-id';
import { ID } from 'types/configuration';
import { ElectionInfo, LightElectionInfo, STATUS } from 'types/electionInfo';
import { unmarshalConfig } from 'types/JSONparser';
import {
  mockElection1,
  mockElection2,
  mockElectionResult21,
  mockElectionResult22,
  mockElectionResult23,
} from './mockData';

const setupMockElection = (): Map<ID, ElectionInfo> => {
  const mockElections: Map<ID, ElectionInfo> = new Map();
  const uid = new ShortUniqueId({ length: 8 });

  const electionID1 = uid();
  const electionID2 = uid();

  mockElections.set(electionID1, {
    ElectionID: electionID1,
    Status: STATUS.OPEN,
    Pubkey: 'XL4V6EMIICW',
    Result: [],
    Configuration: unmarshalConfig(mockElection1),
    BallotSize: 174,
    ChunksPerBallot: 6,
  });

  mockElections.set(electionID2, {
    ElectionID: electionID2,
    Status: STATUS.RESULT_AVAILABLE,
    Pubkey: 'XL4V6EMIICW',
    Result: [mockElectionResult21, mockElectionResult22, mockElectionResult23],
    Configuration: unmarshalConfig(mockElection2),
    BallotSize: 174,
    ChunksPerBallot: 6,
  });

  return mockElections;
};

const toLightElectionInfo = (
  mockElections: Map<ID, ElectionInfo>,
  electionID: ID
): LightElectionInfo => {
  const election = mockElections.get(electionID);

  return {
    ElectionID: electionID,
    Title: election.Configuration.MainTitle,
    Status: election.Status,
    Pubkey: election.Pubkey,
  };
};

export { setupMockElection, toLightElectionInfo };
