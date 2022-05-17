import { ID } from 'types/configuration';
import { ElectionInfo, LightElectionInfo, Results, Status } from 'types/election';
import { unmarshalConfig } from 'types/JSONparser';
import { NodeStatus } from 'types/node';
import {
  mockElection1,
  mockElection2,
  mockElectionResult11,
  mockElectionResult12,
  mockElectionResult21,
  mockElectionResult22,
  mockElectionResult23,
  mockRoster,
} from './mockData';

const setupMockElection = () => {
  const mockElections: Map<ID, ElectionInfo> = new Map();
  const mockResults: Map<ID, Results[]> = new Map();

  // Mock of the DKGStatuses
  const mockDKG: Map<ID, Map<string, NodeStatus>> = new Map();
  const mockDKGSetup: Map<string, NodeStatus> = new Map();
  const mockDKGNotInitialized: Map<string, NodeStatus> = new Map();

  // Mock of the node proxy mapping
  const mockNodeProxyAddresses: Map<ID, Map<string, string>> = new Map();
  const mockAddresses: Map<string, string> = new Map();

  mockRoster.forEach((node) => {
    mockAddresses.set(node, node);
    mockDKGSetup.set(node, NodeStatus.Initialized);
    mockDKGNotInitialized.set(node, NodeStatus.NotInitialized);
  });

  mockDKGSetup.set(mockRoster[0], NodeStatus.Setup);

  const electionID1 = '36kSJ0tH';
  const electionID2 = 'Bnq9gLmf';

  mockElections.set(electionID1, {
    ElectionID: electionID1,
    Status: Status.Initial,
    Pubkey: 'XL4V6EMIICW',
    Result: [],
    Roster: mockRoster,
    Configuration: unmarshalConfig(mockElection1),
    BallotSize: 174,
    ChunksPerBallot: 6,
  });

  mockResults.set(electionID1, [mockElectionResult11, mockElectionResult12]);

  mockDKG.set(electionID1, mockDKGNotInitialized);
  mockNodeProxyAddresses.set(electionID1, mockAddresses);

  mockElections.set(electionID2, {
    ElectionID: electionID2,
    Status: Status.ResultAvailable,
    Pubkey: 'XL4V6EMIICW',
    Result: [],
    Roster: mockRoster,
    Configuration: unmarshalConfig(mockElection2),
    BallotSize: 174,
    ChunksPerBallot: 6,
  });

  mockResults.set(electionID2, [mockElectionResult21, mockElectionResult22, mockElectionResult23]);
  mockDKG.set(electionID2, mockDKGSetup);
  mockNodeProxyAddresses.set(electionID2, mockAddresses);

  for (let j = 0; j < 5; j++) {
    let electionID11 = '36kSJ0t' + (j as number);
    let electionID22 = 'Bnq9gLm' + (j as number);

    mockElections.set(electionID11, {
      ElectionID: electionID11,
      Status: j as Status,
      Pubkey: 'XL4V6EMIICW',
      Result: [],
      Roster: mockRoster,
      Configuration: unmarshalConfig(mockElection1),
      BallotSize: 174,
      ChunksPerBallot: 6,
    });

    mockResults.set(electionID11, [mockElectionResult11, mockElectionResult12]);
    mockNodeProxyAddresses.set(electionID11, mockAddresses);

    mockElections.set(electionID22, {
      ElectionID: electionID22,
      Status: j as Status,
      Pubkey: 'XL4V6EMIICW',
      Result: [],
      Roster: mockRoster,
      Configuration: unmarshalConfig(mockElection2),
      BallotSize: 174,
      ChunksPerBallot: 6,
    });

    mockResults.set(electionID22, [
      mockElectionResult21,
      mockElectionResult22,
      mockElectionResult23,
    ]);
    mockNodeProxyAddresses.set(electionID22, mockAddresses);

    if (j >= Status.Open) {
      mockDKG.set(electionID11, mockDKGSetup);
      mockDKG.set(electionID22, mockDKGSetup);
    } else {
      mockDKG.set(electionID11, mockDKGNotInitialized);
      mockDKG.set(electionID22, mockDKGNotInitialized);
    }
  }

  return { mockElections, mockResults, mockDKG, mockNodeProxyAddresses };
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
