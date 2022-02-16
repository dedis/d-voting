import { rest } from 'msw';

export const handlers = [
  // rest.post('/evoting/', (req, res, ctx) => {
  //   // Persist user's authentication in the session
  //   sessionStorage.setItem('is-authenticated', 'true');

  //   return res(
  //     // Respond with a 200 status code
  //     ctx.status(200)
  //   );
  // }),

  rest.post('/evoting/all', (req, res, ctx) => {
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
