import { ID } from 'types/configuration';
import { ElectionInfo, LightElectionInfo, NodeStatus, Results, Status } from 'types/election';
import { unmarshalConfig } from 'types/JSONparser';
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
  // NodeStatus contains the current status of the nodes, the boolean is set to
  // true if the handler must respond with the updated value (as the handler
  // cannot know when we poll if we have already started setting up or not)
  const mockDKG: Map<ID, [NodeStatus, boolean]> = new Map();

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

  mockDKG.set(electionID1, [NodeStatus.NotInitialized, false]);

  mockElections.set(electionID2, {
    ElectionID: electionID2,
    Status: Status.ResultAvailable,
    Pubkey: 'XL4V6EMIICW',
    Result: [mockElectionResult21, mockElectionResult22, mockElectionResult23],
    Roster: mockRoster,
    Configuration: unmarshalConfig(mockElection2),
    BallotSize: 174,
    ChunksPerBallot: 6,
  });

  mockResults.set(electionID2, [mockElectionResult21, mockElectionResult22, mockElectionResult23]);
  mockDKG.set(electionID2, [NodeStatus.Initialized, true]);

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

    mockElections.set(electionID22, {
      ElectionID: electionID22,
      Status: j as Status,
      Pubkey: 'XL4V6EMIICW',
      Result: [mockElectionResult21, mockElectionResult22, mockElectionResult23],
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

    if (j >= Status.Open) {
      mockDKG.set(electionID11, [NodeStatus.Initialized, true]);
      mockDKG.set(electionID22, [NodeStatus.Initialized, true]);
    } else {
      mockDKG.set(electionID11, [NodeStatus.NotInitialized, false]);
      mockDKG.set(electionID22, [NodeStatus.NotInitialized, false]);
    }
  }

  return { mockElections, mockResults, mockDKG };
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
