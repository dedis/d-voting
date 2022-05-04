import { rest } from 'msw';
import ShortUniqueId from 'short-unique-id';
import { ROUTE_LOGGED } from 'Routes';

import {
  ENDPOINT_GET_TEQ_KEY,
  ENDPOINT_LOGOUT,
  ENDPOINT_PERSONAL_INFO,
} from '../components/utils/Endpoints';
import * as endpoints from '../components/utils/Endpoints';

import {
  EditDKGActorBody,
  EditElectionBody,
  NewDKGBody,
  NewElectionBody,
  NewElectionVoteBody,
  NewUserRole,
  RemoveUserRole,
} from '../types/frontendRequestBody';

import { ID } from 'types/configuration';
import { ACTION, STATUS } from 'types/election';
import { setupMockElection, toLightElectionInfo } from './setupMockElections';
import setupMockUserDB from './setupMockUserDB';
import { ROLE } from 'types/userRole';
import { mockRoster } from './mockData';

const uid = new ShortUniqueId({ length: 8 });
const mockUserID = 561934;

const { mockElections, mockResults } = setupMockElection();

var mockUserDB = setupMockUserDB();

const SETUP_TIMER = 2000;
const SHUFFLE_TIMER = 2000;
const DECRYPT_TIMER = 8000;

export const handlers = [
  rest.get(ENDPOINT_PERSONAL_INFO, (req, res, ctx) => {
    const isLogged = sessionStorage.getItem('is-authenticated') === 'true';
    const userId = isLogged ? mockUserID : 0;
    const userInfos = isLogged
      ? {
          lastname: 'Bobster',
          firstname: 'Alice',
          role: ROLE.Admin,
          sciper: userId,
        }
      : {};

    return res(
      ctx.status(200),
      ctx.json({
        islogged: isLogged,
        ...userInfos,
      })
    );
  }),

  rest.get(ENDPOINT_GET_TEQ_KEY, (req, res, ctx) => {
    const url = ROUTE_LOGGED;
    sessionStorage.setItem('is-authenticated', 'true');
    sessionStorage.setItem('id', '283205');
    return res(ctx.status(200), ctx.json({ url: url }));
  }),

  rest.post(ENDPOINT_LOGOUT, (req, res, ctx) => {
    sessionStorage.setItem('is-authenticated', 'false');
    return res(ctx.status(200));
  }),

  rest.get(endpoints.elections, async (req, res, ctx) => {
    mockElections.forEach(async (election) => {
      switch (election.Status) {
        case STATUS.OnGoingShuffle:
          setTimeout(() => {
            mockElections.set(election.ElectionID as string, {
              ...mockElections.get(election.ElectionID as string),
              Result: mockResults.get(election.ElectionID as string),
              Status: STATUS.ShuffledBallots,
            });
          }, SHUFFLE_TIMER);
          break;
        default:
          break;
      }
    });

    return res(
      ctx.status(200),
      ctx.json({
        Elections: Array.from(mockElections.values()).map((election) =>
          toLightElectionInfo(mockElections, election.ElectionID)
        ),
      })
    );
  }),

  rest.get(endpoints.election(':ElectionID'), async (req, res, ctx) => {
    const { ElectionID } = req.params;

    const mockElection = mockElections.get(ElectionID as string);

    switch (mockElection.Status) {
      case STATUS.OnGoingShuffle:
        setTimeout(() => {
          mockElections.set(ElectionID as string, {
            ...mockElections.get(ElectionID as string),
            Result: mockResults.get(ElectionID as string),
            Status: STATUS.ShuffledBallots,
          });
        }, SHUFFLE_TIMER);
        break;
      default:
        break;
    }

    return res(ctx.status(200), ctx.json(mockElections.get(ElectionID as ID)));
  }),

  rest.post(endpoints.newElection, (req, res, ctx) => {
    const body = req.body as NewElectionBody;

    const createElection = (configuration: any) => {
      const newElectionID = uid();

      mockElections.set(newElectionID, {
        ElectionID: newElectionID,
        Status: STATUS.Initial,
        Pubkey: 'DEAEV6EMII',
        Result: [],
        Roster: mockRoster,
        Configuration: configuration,
        BallotSize: 290,
        ChunksPerBallot: 10,
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

  rest.post(endpoints.newElectionVote(':ElectionID'), (req, res, ctx) => {
    const { Ballot }: NewElectionVoteBody = JSON.parse(req.body.toString());

    return res(
      ctx.status(200),
      ctx.json({
        Ballot: Ballot,
      })
    );
  }),

  rest.put(endpoints.editElection(':ElectionID'), (req, res, ctx) => {
    const body = req.body as EditElectionBody;
    const { ElectionID } = req.params;
    var Status = STATUS.Initial;

    switch (body.Action) {
      case ACTION.Open:
        Status = STATUS.Open;
        break;
      case ACTION.Close:
        Status = STATUS.Closed;
        break;
      case ACTION.CombineShares:
        Status = STATUS.DecryptedBallots;
        break;
      case ACTION.Cancel:
        Status = STATUS.Canceled;
        break;
      default:
        break;
    }
    mockElections.set(ElectionID as string, {
      ...mockElections.get(ElectionID as string),
      Status,
    });

    return res(ctx.status(200), ctx.text('Action successfully done'));
  }),

  rest.put(endpoints.editShuffle(':ElectionID'), (req, res, ctx) => {
    const { ElectionID } = req.params;
    mockElections.set(ElectionID as string, {
      ...mockElections.get(ElectionID as string),
      Status: STATUS.OnGoingShuffle,
    });

    return res(ctx.status(200), ctx.text('Action successfully done'));
  }),

  rest.put(endpoints.editDKGActors(':ElectionID'), (req, res, ctx) => {
    const { ElectionID } = req.params;
    const body = req.body as EditDKGActorBody;
    var status = STATUS.InitializedNodes;

    switch (body.Action) {
      case ACTION.Setup:
        status = STATUS.OnGoingSetup;
        break;
      case ACTION.BeginDecryption:
        status = STATUS.OnGoingDecryption;
        break;
      default:
        break;
    }

    mockElections.set(ElectionID as string, {
      ...mockElections.get(ElectionID as string),
      Result: mockResults.get(ElectionID as string),
      Status: status,
    });

    return res(ctx.status(200), ctx.text('Action successfully done'));
  }),

  rest.get(endpoints.editDKGActors(':ElectionID'), async (req, res, ctx) => {
    const { ElectionID } = req.params;
    const election = mockElections.get(ElectionID as string);

    switch (election.Status) {
      case STATUS.OnGoingSetup:
        setTimeout(() => {
          mockElections.set(ElectionID as string, {
            ...mockElections.get(ElectionID as string),
            Result: mockResults.get(ElectionID as string),
            Status: STATUS.Setup,
          });

          console.log('updated Setup status');
        }, SETUP_TIMER);
        break;
      case STATUS.OnGoingDecryption:
        setTimeout(() => {
          mockElections.set(ElectionID as string, {
            ...mockElections.get(ElectionID as string),
            Result: mockResults.get(ElectionID as string),
            Status: STATUS.DecryptedBallots,
          });

          console.log('updated Decrypted status');
        }, DECRYPT_TIMER);
        break;
      default:
        break;
    }

    return res(ctx.status(200), ctx.json({ Status: election.Status }));
  }),

  rest.post(endpoints.dkgActors, (req, res, ctx) => {
    const body = req.body as NewDKGBody;

    return res(ctx.status(200));
  }),

  rest.get(endpoints.ENDPOINT_USER_RIGHTS, (req, res, ctx) => {
    return res(ctx.status(200), ctx.json(mockUserDB.filter((user) => user.role !== 'voter')));
  }),

  rest.post(endpoints.ENDPOINT_ADD_ROLE, (req, res, ctx) => {
    const body = req.body as NewUserRole;
    mockUserDB.push({ id: uid(), ...body });

    return res(ctx.status(200));
  }),

  rest.post(endpoints.ENDPOINT_REMOVE_ROLE, (req, res, ctx) => {
    const body = req.body as RemoveUserRole;
    mockUserDB = mockUserDB.filter((user) => user.sciper !== body.sciper);

    return res(ctx.status(200));
  }),
];
