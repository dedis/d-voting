import { rest } from 'msw';

import { CREATE_ENDPOINT, GET_ALL_ELECTIONS_ENDPOINT } from '../components/utils/Endpoints';
import { GET_PERSONNAL_INFOS } from '../components/utils/ExpressEndoints';

export const handlers = [
  rest.get(GET_PERSONNAL_INFOS, (req, res, ctx) => {
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

  rest.get('/api/getTkKey', (req, res, ctx) => {
    // const key = 'abcdef123456';
    // const url = `/mocked-tequilla/requestauth?requestkey=${key}`;
    const url = '/';
    sessionStorage.setItem('is-authenticated', 'true');

    return res(ctx.status(200), ctx.json({ url: url }));
  }),

  // rest.get('/api/mocked-tequilla/requestauth', (req, res, ctx) => {
  //   return res(ctx.status(200));
  // }),

  // rest.get('/api/control_key', (req, res, ctx) => {
  //   return res(
  //     ctx.status(200),
  //     ctx.json({
  //       role: 'voter',
  //     })
  //   );
  // }),

  rest.post('/api/logout', (req, res, ctx) => {
    sessionStorage.setItem('is-authenticated', 'false');
    return res(ctx.status(200));
  }),

  rest.post(GET_ALL_ELECTIONS_ENDPOINT, (req, res, ctx) => {
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

  rest.post(CREATE_ENDPOINT, (req, res, ctx) => {
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
];
