import { ID } from 'types/configuration';
import { FormInfo, LightFormInfo, Results, Status } from 'types/form';
import { unmarshalConfig } from 'types/JSONparser';
import { NodeStatus } from 'types/node';
import {
  mockForm1,
  mockForm2,
  mockForm3,
  mockFormResult11,
  mockFormResult12,
  mockFormResult21,
  mockFormResult22,
  mockFormResult23,
  mockFormResult31,
  mockFormResult32,
  mockFormResult33,
  mockNodes,
  mockRoster,
} from './mockData';

const setupMockForm = () => {
  const mockForms: Map<ID, FormInfo> = new Map();
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

  const formID1 = '36kSJ0tH';
  const formID2 = 'Bnq9gLmf';
  const formID3 = 'Afdv4ffl';

  mockForms.set(formID1, {
    FormID: formID1,
    Status: Status.Initial,
    Pubkey: 'XL4V6EMIICW',
    Result: [],
    Roster: mockRoster,
    Configuration: unmarshalConfig(mockForm1),
    BallotSize: 174,
    ChunksPerBallot: 6,
    Voters: [],
  });

  mockResults.set(formID1, [mockFormResult11, mockFormResult12]);

  mockDKG.set(formID1, mockDKGNotInitialized);

  mockForms.set(formID2, {
    FormID: formID2,
    Status: Status.ResultAvailable,
    Pubkey: 'XL4V6EMIICW',
    Result: [mockFormResult21, mockFormResult22, mockFormResult23],
    Roster: mockRoster,
    Configuration: unmarshalConfig(mockForm2),
    BallotSize: 174,
    ChunksPerBallot: 6,
    Voters: ['aefae', 'ozeivn', 'ovaeop'],
  });

  mockResults.set(formID2, [mockFormResult21, mockFormResult22, mockFormResult23]);
  mockDKG.set(formID2, mockDKGSetup);

  mockForms.set(formID3, {
    FormID: formID3,
    Status: Status.Open,
    Pubkey: 'XL4V6EMIICW',
    Result: [],
    Roster: mockRoster,
    Configuration: unmarshalConfig(mockForm3),
    BallotSize: 291,
    ChunksPerBallot: 11,
    Voters: [],
  });

  mockResults.set(formID3, [mockFormResult31, mockFormResult32, mockFormResult33]);
  mockDKG.set(formID3, mockDKGSetup);

  return { mockForms, mockResults, mockDKG, mockNodeProxyAddresses };
};

const toLightFormInfo = (mockForms: Map<ID, FormInfo>, formID: ID): LightFormInfo => {
  const form = mockForms.get(formID);

  return {
    FormID: formID,
    Title: form.Configuration.Title,
    Status: form.Status,
    Pubkey: form.Pubkey,
  };
};

export { setupMockForm, toLightFormInfo };
