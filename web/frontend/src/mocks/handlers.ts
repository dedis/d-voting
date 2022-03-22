import { rest } from 'msw';
import { mockElection1, mockElection2 } from './mockData';

import {
  ENDPOINT_EVOTING_CAST_BALLOT,
  ENDPOINT_EVOTING_GET_ALL,
  ENDPOINT_EVOTING_GET_ELECTION,
  ENDPOINT_GET_TEQ_KEY,
  ENDPOINT_LOGOUT,
  ENDPOINT_PERSONAL_INFO,
} from '../components/utils/Endpoints';

var mockUserID = 561934;

interface GetElectionBody {
  ElectionID: string;
  Token: string;
}

var mockElections = [
  {
    ElectionID: 'election ID1',
    Title: 'Title Election 1',
    Status: 1,
    Pubkey: 'XL4V6EMIICW',
    Result: [],
    Format: mockElection1,
  },
  {
    ElectionID: 'election ID2',
    Title: 'Title Election 2',
    Status: 1,
    Pubkey: 'XL4V6EMIICW',
    Result: [],
    Format: mockElection2,
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
    const url = '/';
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

  rest.post(ENDPOINT_EVOTING_CAST_BALLOT, (req, res, ctx) => {
    const body: GetElectionBody = JSON.parse(req.body.toString());
    console.log(body);

    return res(
      ctx.status(200),
      ctx.json({
        ElectionID: body.ElectionID,
        UserID: mockUserID,
        Ballot: {
          K: 'KOALA',
          C: 'CHARLIE',
        },
        Token: body.Token,
      })
    );
  }),
];
