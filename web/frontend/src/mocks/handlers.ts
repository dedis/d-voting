import { rest } from 'msw';
import { unmarshalConfig } from 'types/JSONparser';
import ShortUniqueId from 'short-unique-id';
import { ROUTE_LOGGED } from 'Routes';

import {
  ENDPOINT_EVOTING_CAST_BALLOT,
  ENDPOINT_EVOTING_CREATE,
  ENDPOINT_EVOTING_GET_ALL,
  ENDPOINT_EVOTING_GET_ELECTION,
  ENDPOINT_GET_TEQ_KEY,
  ENDPOINT_LOGOUT,
  ENDPOINT_PERSONAL_INFO,
} from '../components/utils/Endpoints';

import {
  CreateElectionBody,
  CreateElectionCastVote,
  GetElectionBody,
} from '../types/frontendRequestBody';
import { mockElection1, mockElection2 } from './mockData';

const uid = new ShortUniqueId({ length: 8 });
const mockUserID = 561934;

var mockElections = [
  {
    ElectionID: uid(),
    Title: 'Title Election 1',
    Status: 1,
    Pubkey: 'XL4V6EMIICW',
    Result: [],
    Format: unmarshalConfig(mockElection1),
    BallotSize: 174,
    ChunksPerBallot: 6,
  },
  {
    ElectionID: uid(),
    Title: 'Title Election 2',
    Status: 1,
    Pubkey: 'XL4V6EMIICW',
    Result: [],
    Format: unmarshalConfig(mockElection2),
    BallotSize: 174,
    ChunksPerBallot: 6,
  },
];

export const handlers = [
  rest.get(ENDPOINT_PERSONAL_INFO, (req, res, ctx) => {
    let isLogged = sessionStorage.getItem('is-authenticated') === 'true';
    let userId = isLogged ? mockUserID : 0;

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

    return res(ctx.status(200), ctx.json({ url: url }));
  }),

  rest.post(ENDPOINT_LOGOUT, (req, res, ctx) => {
    sessionStorage.setItem('is-authenticated', 'false');
    return res(ctx.status(200));
  }),

  rest.get(ENDPOINT_EVOTING_GET_ALL, (req, res, ctx) => {
    return res(
      ctx.status(200),
      ctx.json({
        AllElectionsInfo: mockElections,
      })
    );
  }),

  rest.post(ENDPOINT_EVOTING_GET_ALL, (req, res, ctx) => {
    return res(
      ctx.status(200),
      ctx.json({
        AllElectionsInfo: mockElections,
      })
    );
  }),

  rest.post(ENDPOINT_EVOTING_GET_ELECTION, (req, res, ctx) => {
    const body: GetElectionBody = JSON.parse(req.body.toString());

    return res(
      ctx.status(200),
      ctx.json(mockElections.find((election) => election.ElectionID === body.ElectionID))
    );
  }),

  rest.post(ENDPOINT_EVOTING_CREATE, (req, res, ctx) => {
    const body: CreateElectionBody = JSON.parse(req.body.toString());

    const createElection = (format: any) => {
      const newElection = {
        ElectionID: uid(),
        Title: format.MainTitle,
        Status: 1,
        Pubkey: 'DEAEV6EMII',
        Result: [],
        Format: format,
        BallotSize: 290,
        ChunksPerBallot: 10,
      };
      mockElections.push(newElection);
      return newElection.ElectionID;
    };

    return res(
      ctx.status(200),
      ctx.json({
        ElectionID: createElection(body.Format),
      })
    );
  }),

  rest.post(ENDPOINT_EVOTING_CAST_BALLOT, (req, res, ctx) => {
    const body: CreateElectionCastVote = JSON.parse(req.body.toString());

    return res(
      ctx.status(200),
      ctx.json({
        UserID: body.UserID,
        Ballot: body.Ballot,
      })
    );
  }),
];
