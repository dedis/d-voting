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
  CANCELED_STATUS,
  CLOSED_STATUS,
  CreateElectionBody,
  CreateElectionCastVote,
  DECRYPTED_BALLOTS_STATUS,
  ElectionActionsBody,
  INITIAL_STATUS,
  OPEN_STATUS,
  RESULT_AVAILABLE_STATUS,
  SHUFFLED_BALLOTS_STATUS,
} from '../types/frontendRequestBody';
import { mockElection1, mockElection2 } from './mockData';

const uid = new ShortUniqueId({ length: 8 });
const mockUserID = 561934;

var mockElections = [
  {
    ElectionID: uid(),
    Title: 'Title Election 1',
    Status: OPEN_STATUS,
    Pubkey: 'XL4V6EMIICW',
    Result: [],
    Configuration: unmarshalConfig(mockElection1),
    BallotSize: 174,
    ChunksPerBallot: 6,
  },
  {
    ElectionID: uid(),
    Title: 'Title Election 2',
    Status: OPEN_STATUS,
    Pubkey: 'XL4V6EMIICW',
    Result: [],
    Configuration: unmarshalConfig(mockElection2),
    BallotSize: 174,
    ChunksPerBallot: 6,
  },
];

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
    // TODO: GET ALL SHOULD ONLY RETURN SOME FIELDS OF THE ELECTION BEFORE
    // ADAPTING THE MOCK, THE FRONTEND SHOULD BE UPDATED TO ACCEPT ONLY THESE
    // FIELDS const Elections = mockElections.map(({ ElectionID, Title, Status,
    // Pubkey }) => ({ ElectionID, Title, Status, Pubkey,
    // }));
    return res(
      ctx.status(200),
      ctx.json({
        Elections: mockElections,
      })
    );
  }),

  rest.get(ENDPOINT_EVOTING_GET_ELECTION(), (req, res, ctx) => {
    const { ElectionID } = req.params;
    return res(
      ctx.status(200),
      ctx.json(mockElections.find((election) => election.ElectionID === ElectionID))
    );
  }),

  rest.post(ENDPOINT_EVOTING_CREATE, (req, res, ctx) => {
    const body: CreateElectionBody = JSON.parse(req.body.toString());

    const createElection = (Configuration: any) => {
      const newElection = {
        ElectionID: uid(),
        Title: Configuration.MainTitle,
        Status: 1,
        Pubkey: 'DEAEV6EMII',
        Result: [],
        Configuration: Configuration,
        BallotSize: 290,
        ChunksPerBallot: 10,
      };
      mockElections.push(newElection);
      return newElection.ElectionID;
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
    var Status = INITIAL_STATUS;
    const foundIndex = mockElections.findIndex((x) => x.ElectionID === ElectionID);
    switch (body.Action) {
      case 'open':
        Status = OPEN_STATUS;
        break;
      case 'close':
        Status = CLOSED_STATUS;
        break;
      case 'combineShares':
        Status = DECRYPTED_BALLOTS_STATUS;
        break;
      case 'cancel':
        Status = CANCELED_STATUS;
        break;
      default:
        break;
    }
    mockElections[foundIndex] = { ...mockElections[foundIndex], Status };
    return res(ctx.status(200), ctx.text('Action sucessfully done'));
  }),

  rest.put(ENDPOINT_EVOTING_SHUFFLE(), (req, res, ctx) => {
    const { ElectionID } = req.params;
    const foundIndex = mockElections.findIndex((x) => x.ElectionID === ElectionID);
    mockElections[foundIndex] = { ...mockElections[foundIndex], Status: SHUFFLED_BALLOTS_STATUS };
    return res(ctx.status(200), ctx.text('Action sucessfully done'));
  }),

  rest.put(ENDPOINT_EVOTING_DECRYPT(), (req, res, ctx) => {
    const { ElectionID } = req.params;
    const foundIndex = mockElections.findIndex((x) => x.ElectionID === ElectionID);
    mockElections[foundIndex] = { ...mockElections[foundIndex], Status: RESULT_AVAILABLE_STATUS };
    return res(ctx.status(200), ctx.text('Action sucessfully done'));
  }),
];
