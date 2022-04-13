import { rest } from 'msw';
import { unmarshalConfig } from 'types/JSONparser';
import ShortUniqueId from 'short-unique-id';
import { ROUTE_LOGGED } from 'Routes';

import {
  ENDPOINT_EVOTING_CAST_BALLOT,
  ENDPOINT_EVOTING_CREATE,
  ENDPOINT_EVOTING_DECRYPT,
  ENDPOINT_EVOTING_ELECTION,
  ENDPOINT_EVOTING_GET_ALL,
  ENDPOINT_EVOTING_GET_ELECTION,
  ENDPOINT_EVOTING_SHUFFLE,
  ENDPOINT_GET_TEQ_KEY,
  ENDPOINT_LOGOUT,
  ENDPOINT_PERSONAL_INFO,
} from '../components/utils/Endpoints';

import {
  CreateElectionBody,
  CreateElectionCastVote,
  ElectionInfo,
  LightElectionInfo,
  ElectionActionsBody,
} from '../types/frontendRequestBody';
import { mockElection1, mockElection2 } from './mockData';
import { ID } from 'types/configuration';

const uid = new ShortUniqueId({ length: 8 });
const mockUserID = 561934;

const mockElections: Map<ID, ElectionInfo> = new Map();
const mockElectionsLight: Map<ID, LightElectionInfo> = new Map();
const electionID1 = uid();
const electionID2 = uid();

mockElections.set(electionID1, {
  ElectionID: electionID1,
  Status: 1,
  Pubkey: 'XL4V6EMIICW',
  Result: [],
  Configuration: unmarshalConfig(mockElection1),
  BallotSize: 174,
  ChunksPerBallot: 6,
});
mockElections.set(electionID2, {
  ElectionID: electionID2,
  Status: 1,
  Pubkey: 'XL4V6EMIICW',
  Result: [],
  Configuration: unmarshalConfig(mockElection2),
  BallotSize: 174,
  ChunksPerBallot: 6,
});

mockElectionsLight.set(electionID1, {
  ElectionID: electionID1,
  Title: unmarshalConfig(mockElection1).MainTitle,
  Status: 1,
  Pubkey: 'XL4V6EMIICW',
});

mockElectionsLight.set(electionID2, {
  ElectionID: electionID2,
  Title: unmarshalConfig(mockElection2).MainTitle,
  Status: 1,
  Pubkey: 'XL4V6EMIICW',
});

export const handlers = [
  rest.get(ENDPOINT_PERSONAL_INFO, (req, res, ctx) => {
    const isLogged = sessionStorage.getItem('is-authenticated') === 'true';
    const userId = isLogged ? mockUserID : 0;

    return res(
      ctx.status(200),
      ctx.json({
        islogged: isLogged,
        lastname: 'Bobster',
        firstname: 'Alice',
        role: 'admin',
        sciper: userId,
      })
    );
  }),

  rest.get(ENDPOINT_GET_TEQ_KEY, (req, res, ctx) => {
    const url = ROUTE_LOGGED;
    sessionStorage.setItem('is-authenticated', 'true');
    sessionStorage.setItem('id', '283205');
    sessionStorage.setItem('token', 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9');
    return res(ctx.status(200), ctx.json({ url: url }));
  }),

  rest.post(ENDPOINT_LOGOUT, (req, res, ctx) => {
    sessionStorage.setItem('is-authenticated', 'false');
    return res(ctx.status(200));
  }),

  rest.get(ENDPOINT_EVOTING_GET_ALL, (req, res, ctx) => {
    // TODO: GET GET ALL SHOULD ONLY RETURN SOME FIELDS OF THE ELECTION
    // BEFORE ADAPTING THE MOCK, THE FRONTEND SHOULD BE UPDATED TO ACCEPT THIS
    // const Elections = mockElections.map(({ ElectionID, Title, Status, Pubkey }) => ({
    //   ElectionID,
    //   Title,
    //   Status,
    //   Pubkey,
    // }));
    return res(
      ctx.status(200),
      ctx.json({
        AllElectionsInfo: Array.from(mockElectionsLight.values()),
      })
    );
  }),

  rest.get(ENDPOINT_EVOTING_GET_ELECTION(), (req, res, ctx) => {
    const electionID = req.url.toString().split('/').at(-1);
    const { ElectionID } = req.params;

    return res(ctx.status(200), ctx.json(mockElections.get(ElectionID as ID)));
  }),

  rest.post(ENDPOINT_EVOTING_CREATE, (req, res, ctx) => {
    const body: CreateElectionBody = JSON.parse(req.body.toString());

    const createElection = (configuration: any) => {
      const newElectionID = uid();

      mockElections.set(newElectionID, {
        ElectionID: newElectionID,
        Status: 1,
        Pubkey: 'DEAEV6EMII',
        Result: [],
        Configuration: configuration,
        BallotSize: 290,
        ChunksPerBallot: 10,
      });
      mockElectionsLight.set(newElectionID, {
        ElectionID: newElectionID,
        Title: configuration.MainTitle,
        Status: 1,
        Pubkey: 'DEAEV6EMII',
      });

      return newElectionID;
    };

    return res(
      ctx.status(200),
      ctx.json({
        ElectionID: createElection(body.Configuration),
      })
    );
  }),

  rest.post(ENDPOINT_EVOTING_CAST_BALLOT(), (req, res, ctx) => {
    const { Ballot }: CreateElectionCastVote = JSON.parse(req.body.toString());

    return res(
      ctx.status(200),
      ctx.json({
        Ballot: Ballot,
      })
    );
  }),

  rest.put(ENDPOINT_EVOTING_ELECTION(), (req, res, ctx) => {
    const body: ElectionActionsBody = JSON.parse(req.body.toString());
    const { ElectionID } = req.params;
    var Status = 1;
    switch (body.Action) {
      case 'open':
        Status = 1;
        break;
      case 'close':
        Status = 2;
        break;
      case 'combineShares':
        Status = 4;
        break;
      case 'cancel':
        Status = 6;
        break;
      default:
        break;
    }
    mockElections.set(ElectionID as string, {
      ...mockElections.get(ElectionID as string),
      Status,
    });
    mockElectionsLight.set(ElectionID as string, {
      ...mockElectionsLight.get(ElectionID as string),
      Status,
    });
    return res(ctx.status(200), ctx.text('Action successfully done'));
  }),

  rest.put(ENDPOINT_EVOTING_SHUFFLE(), (req, res, ctx) => {
    const { ElectionID } = req.params;
    var Status = 3;
    mockElections.set(ElectionID as string, {
      ...mockElections.get(ElectionID as string),
      Status,
    });
    mockElectionsLight.set(ElectionID as string, {
      ...mockElectionsLight.get(ElectionID as string),
      Status,
    });
    return res(ctx.status(200), ctx.text('Action successfully done'));
  }),

  rest.put(ENDPOINT_EVOTING_DECRYPT(), (req, res, ctx) => {
    const { ElectionID } = req.params;
    var Status = 5;
    mockElections.set(ElectionID as string, {
      ...mockElections.get(ElectionID as string),
      Status,
    });
    mockElectionsLight.set(ElectionID as string, {
      ...mockElectionsLight.get(ElectionID as string),
      Status,
    });
    return res(ctx.status(200), ctx.text('Action successfully done'));
  }),
];
