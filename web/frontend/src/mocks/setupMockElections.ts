import { ID } from 'types/configuration';
import { ElectionInfo, LightElectionInfo, Results, Status } from 'types/election';
import { unmarshalConfig } from 'types/JSONparser';
import { NodeStatus } from 'types/node';
import {
  mockElection1,
  mockElection2,
  mockElection3,
  mockElectionResult11,
  mockElectionResult12,
  mockElectionResult21,
  mockElectionResult22,
  mockElectionResult23,
  mockElectionResult31,
  mockElectionResult32,
  mockElectionResult33,
  mockNodes,
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
  const mockNodeProxyAddresses: Map<string, string> = new Map();

  mockNodes.forEach((node, index) => {
    mockNodeProxyAddresses.set(node, 'https://example' + index + '.com');
  });

  mockRoster.forEach((node) => {
    mockDKGSetup.set(node, NodeStatus.Initialized);
    mockDKGNotInitialized.set(node, NodeStatus.NotInitialized);
  });

  mockDKGSetup.set(mockRoster[0], NodeStatus.Setup);

  const electionID1 = '36kSJ0tH';
  const electionID2 = 'Bnq9gLmf';
  const electionID3 = 'Afdv4ffl';

  mockElections.set(electionID1, {
    ElectionID: electionID1,
    Status: Status.Initial,
    Pubkey: 'XL4V6EMIICW',
    Result: [],
    Roster: mockRoster,
    Configuration: unmarshalConfig(mockElection1),
    BallotSize: 174,
    ChunksPerBallot: 6,
    Voters: [],
  });

  mockResults.set(electionID1, [mockElectionResult11, mockElectionResult12]);

  mockDKG.set(electionID1, mockDKGNotInitialized);

  mockElections.set(electionID2, {
    ElectionID: electionID2,
    Status: Status.ResultAvailable,
    Pubkey: 'XL4V6EMIICW',
    Result: [mockElectionResult21, mockElectionResult22, mockElectionResult23],
    Roster: mockRoster,
    Configuration: unmarshalConfig(mockElection2),
    BallotSize: 174,
    ChunksPerBallot: 6,
    Voters: ['aefae', 'ozeivn', 'ovaeop'],
  });

  mockResults.set(electionID2, [mockElectionResult21, mockElectionResult22, mockElectionResult23]);
  mockDKG.set(electionID2, mockDKGSetup);

  mockElections.set(electionID3, {
    ElectionID: electionID3,
    Status: Status.Open,
    Pubkey: 'XL4V6EMIICW',
    Result: [],
    Roster: mockRoster,
    Configuration: unmarshalConfig(mockElection3),
    BallotSize: 291,
    ChunksPerBallot: 11,
    Voters: [],
  });

  mockResults.set(electionID3, [mockElectionResult31, mockElectionResult32, mockElectionResult33]);
  mockDKG.set(electionID3, mockDKGSetup);

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
