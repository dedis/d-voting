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
  EditElectionBody,
  NewElectionBody,
  NewElectionVoteBody,
  NewUserRole,
  RemoveUserRole,
} from '../types/frontendRequestBody';

import { ID } from 'types/configuration';
import { STATUS } from 'types/electionInfo';
import { setupMockElection, toLightElectionInfo } from './setupMockElections';
import setupMockUserDB from './setupMockUserDB';
import { Role } from 'types/userRole';

const uid = new ShortUniqueId({ length: 8 });
const mockUserID = 561934;

const { mockElections, mockResults } = setupMockElection();

var mockUserDB = setupMockUserDB();

export const handlers = [
  rest.get(ENDPOINT_PERSONAL_INFO, (req, res, ctx) => {
    const isLogged = sessionStorage.getItem('is-authenticated') === 'true';
    const userId = isLogged ? mockUserID : 0;
    const userInfos = isLogged
      ? {
          lastname: 'Bobster',
          firstname: 'Alice',
          role: Role.Admin,
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

  rest.get(endpoints.elections, (req, res, ctx) => {
    return res(
      ctx.status(200),
      ctx.json({
        Elections: Array.from(mockElections.values()).map((election) =>
          toLightElectionInfo(mockElections, election.ElectionID)
        ),
      })
    );
  }),

  rest.get(endpoints.election(':ElectionID'), (req, res, ctx) => {
    const { ElectionID } = req.params;

    return res(ctx.status(200), ctx.json(mockElections.get(ElectionID as ID)));
  }),

  rest.post(endpoints.newElection, (req, res, ctx) => {
    const body = req.body as NewElectionBody;

    const createElection = (configuration: any) => {
      const newElectionID = uid();

      mockElections.set(newElectionID, {
        ElectionID: newElectionID,
        Status: STATUS.Open,
        Pubkey: 'DEAEV6EMII',
        Result: [],
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
      case 'open':
        Status = STATUS.Open;
        break;
      case 'close':
        Status = STATUS.Closed;
        break;
      case 'combineShares':
        Status = STATUS.DecryptedBallots;
        break;
      case 'cancel':
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
      Status: STATUS.ShuffledBallots,
    });

    return res(ctx.status(200), ctx.text('Action successfully done'));
  }),

  rest.put(endpoints.editDKGActors(':ElectionID'), (req, res, ctx) => {
    const { ElectionID } = req.params;
    mockElections.set(ElectionID as string, {
      ...mockElections.get(ElectionID as string),
      Result: mockResults.get(ElectionID as string),
      Status: STATUS.ResultAvailable,
    });

    return res(ctx.status(200), ctx.text('Action successfully done'));
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
