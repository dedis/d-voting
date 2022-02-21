import { rest } from 'msw';

import {
  ENDPOINT_GET_TEQ_KEY,
  ENDPOINT_LOGOUT,
  ENDPOINT_PERSONNAL_INFO,
  ENDPOINT_EVOTING_CREATE,
  ENDPOINT_EVOTING_GET_ALL,
} from '../components/utils/Endpoints';

export const handlers = [
  rest.get(ENDPOINT_PERSONNAL_INFO, (req, res, ctx) => {
    let isLogged = sessionStorage.getItem('is-authenticated') === 'true';
    let userId = isLogged ? 561934 : 0;

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
        AllElectionsInfo: [
          {
            ElectionID: 'election ID1',
            Title: 'election TITLE',
            Status: 3,
            Pubkey: 'DEADBEEF',
            Result: [{ Vote: 'vote' }],
            Format: { Candidates: ['candiate1', 'candiate2'] },
          },
        ],
      })
    );
  }),

  rest.post(ENDPOINT_EVOTING_CREATE, (req, res, ctx) => {
    return res(ctx.status(200));
  }),
];
