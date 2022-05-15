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
  NewProxyAddress,
  NewUserRole,
  RemoveUserRole,
} from '../types/frontendRequestBody';

import { ID } from 'types/configuration';
import { Action, Status } from 'types/election';
import { setupMockElection, toLightElectionInfo } from './setupMockElections';
import setupMockUserDB from './setupMockUserDB';
import { UserRole } from 'types/userRole';
import { mockRoster } from './mockData';
import { NodeStatus } from 'types/node';

const uid = new ShortUniqueId({ length: 8 });
const mockUserID = 561934;

const { mockElections, mockResults, mockDKG, mockNodeProxyAddresses } = setupMockElection();

var mockUserDB = setupMockUserDB();

const RESPONSE_TIME = 200;
const CHANGE_STATUS_TIMER = 2000;
const INIT_TIMER = 4000;
const SETUP_TIMER = 4000;
const SHUFFLE_TIMER = 2000;
const DECRYPT_TIMER = 8000;

const isAuthorized = (roles: UserRole[]): boolean => {
  const id = sessionStorage.getItem('id');
  const userRole = mockUserDB.find(({ sciper }) => sciper === id).role;

  if (roles.includes(userRole)) {
    return true;
  }
  return false;
};

export const handlers = [
  rest.get(ENDPOINT_PERSONAL_INFO, async (req, res, ctx) => {
    const isLogged = sessionStorage.getItem('is-authenticated') === 'true';
    const userId = isLogged ? mockUserID : 0;
    const userInfos = isLogged
      ? {
          lastname: 'Bobster',
          firstname: 'Alice',
          role: UserRole.Admin,
          sciper: userId,
        }
      : {};
    await new Promise((r) => setTimeout(r, RESPONSE_TIME));

    return res(
      ctx.status(200),
      ctx.json({
        islogged: isLogged,
        ...userInfos,
      })
    );
  }),

  rest.get(ENDPOINT_GET_TEQ_KEY, async (req, res, ctx) => {
    const url = ROUTE_LOGGED;
    sessionStorage.setItem('is-authenticated', 'true');
    sessionStorage.setItem('id', mockUserID.toString());

    await new Promise((r) => setTimeout(r, RESPONSE_TIME));

    return res(ctx.status(200), ctx.json({ url: url }));
  }),

  rest.post(ENDPOINT_LOGOUT, (req, res, ctx) => {
    sessionStorage.setItem('is-authenticated', 'false');
    return res(ctx.status(200));
  }),

  rest.get(endpoints.elections, async (req, res, ctx) => {
    await new Promise((r) => setTimeout(r, RESPONSE_TIME));

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
    await new Promise((r) => setTimeout(r, RESPONSE_TIME));

    return res(ctx.status(200), ctx.json(mockElections.get(ElectionID as ID)));
  }),

  rest.post(endpoints.newElection, async (req, res, ctx) => {
    const body = req.body as NewElectionBody;

    await new Promise((r) => setTimeout(r, RESPONSE_TIME));

    if (!isAuthorized([UserRole.Admin, UserRole.Operator])) {
      return res(
        ctx.status(403),
        ctx.json({ message: 'You are not authorized to create an election' })
      );
    }

    const createElection = (configuration: any) => {
      const newElectionID = uid();
      mockDKG.set(newElectionID, NodeStatus.NotInitialized);
      const newNodeProxyAddresses = new Map();
      mockRoster.forEach((node) => newNodeProxyAddresses.set(node, node));
      mockNodeProxyAddresses.set(newElectionID, newNodeProxyAddresses);

      mockElections.set(newElectionID, {
        ElectionID: newElectionID,
        Status: Status.Initial,
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

  rest.post(endpoints.newElectionVote(':ElectionID'), async (req, res, ctx) => {
    const { Ballot }: NewElectionVoteBody = req.body as NewElectionVoteBody;
    await new Promise((r) => setTimeout(r, RESPONSE_TIME));

    return res(
      ctx.status(200),
      ctx.json({
        Ballot: Ballot,
      })
    );
  }),

  rest.put(endpoints.editElection(':ElectionID'), async (req, res, ctx) => {
    const body = req.body as EditElectionBody;
    const { ElectionID } = req.params;
    var status = Status.Initial;
    const Result = [];

    await new Promise((r) => setTimeout(r, 1000));

    if (!isAuthorized([UserRole.Admin, UserRole.Operator])) {
      return res(
        ctx.status(403),
        ctx.json({ message: 'You are not authorized to update an election' })
      );
    }

    switch (body.Action) {
      case Action.Open:
        status = Status.Open;
        break;
      case Action.Close:
        status = Status.Closed;
        //return res(ctx.status(500), ctx.text('Oh oh'));

        break;
      case Action.CombineShares:
        status = Status.ResultAvailable;
        mockResults.get(ElectionID as string).forEach((result) => Result.push(result));
        break;
      case Action.Cancel:
        status = Status.Canceled;
        //return res(ctx.status(500), ctx.text('Oh oh'));

        break;
      default:
        break;
    }

    setTimeout(
      () =>
        mockElections.set(ElectionID as string, {
          ...mockElections.get(ElectionID as string),
          Status: status,
          Result,
        }),
      CHANGE_STATUS_TIMER
    );

    return res(ctx.status(200), ctx.text('Action successfully done'));
  }),

  rest.delete(endpoints.editElection(':ElectionID'), async (req, res, ctx) => {
    const { ElectionID } = req.params;
    mockElections.delete(ElectionID as string);
    await new Promise((r) => setTimeout(r, RESPONSE_TIME));

    return res(ctx.status(200), ctx.text('Election deleted'));
  }),

  rest.post(endpoints.dkgActors, (req, res, ctx) => {
    const body = req.body as NewDKGBody;
    setTimeout(() => mockDKG.set(body.ElectionID, NodeStatus.Initialized), INIT_TIMER);

    return res(ctx.status(200));
  }),

  rest.put(endpoints.editDKGActors(':ElectionID'), async (req, res, ctx) => {
    const { ElectionID } = req.params;
    const body = req.body as EditDKGActorBody;

    switch (body.Action) {
      case Action.Setup:
        setTimeout(() => mockDKG.set(ElectionID as string, NodeStatus.Setup), SETUP_TIMER);
        break;
      case Action.BeginDecryption:
        setTimeout(
          () =>
            mockElections.set(ElectionID as string, {
              ...mockElections.get(ElectionID as string),
              Status: Status.PubSharesSubmitted,
            }),
          DECRYPT_TIMER
        );

        break;
      default:
        break;
    }

    return res(ctx.status(200), ctx.text('Action successfully done'));
  }),

  rest.get(endpoints.getDKGActors('*', ':ElectionID'), async (req, res, ctx) => {
    const { ElectionID } = req.params;
    const proxy = req.params[0];
    const election = mockElections.get(ElectionID as string);

    switch (election.Status) {
      case Status.Initial:
        const currentNodeStatus = mockDKG.get(ElectionID as string);

        if (currentNodeStatus === NodeStatus.Initialized) {
          //return res(ctx.status(500), ctx.json('Oh oh'));
          return res(ctx.status(200), ctx.json({ Status: currentNodeStatus, Error: {} }));
        } else if (currentNodeStatus === NodeStatus.Setup) {
          //mock a node setup failure
          /*return res(
            ctx.status(200),
            ctx.json({
              Status: NodeStatus.Failed,
              Error: {
                Title: 'Node could not be initialized',
                Code: '<uint>',
                Message: 'Failed to setup the node',
                Args: {},
              },
            })
          );*/
          return res(ctx.status(200), ctx.json({ Status: currentNodeStatus, Error: {} }));
        } else {
          return res(ctx.status(404), ctx.json(`Election ${ElectionID} does not exist`));
        }
      default:
        return res(ctx.status(200), ctx.json({ Status: NodeStatus.Setup, Error: {} }));
    }
  }),

  rest.put(endpoints.editShuffle(':ElectionID'), async (req, res, ctx) => {
    const { ElectionID } = req.params;

    if (!isAuthorized([UserRole.Admin, UserRole.Operator])) {
      return res(
        ctx.status(403),
        ctx.json({ message: 'You are not authorized to update an election' })
      );
    }

    setTimeout(
      () =>
        mockElections.set(ElectionID as string, {
          ...mockElections.get(ElectionID as string),
          Status: Status.ShuffledBallots,
        }),
      SHUFFLE_TIMER
    );

    return res(ctx.status(200), ctx.text('Action successfully done'));
  }),

  rest.get(endpoints.ENDPOINT_USER_RIGHTS, async (req, res, ctx) => {
    await new Promise((r) => setTimeout(r, RESPONSE_TIME));

    if (!isAuthorized([UserRole.Admin])) {
      return res(
        ctx.status(403),
        ctx.json({ message: 'You are not authorized to get users rights' })
      );
    }

    return res(ctx.status(200), ctx.json(mockUserDB.filter((user) => user.role !== 'voter')));
  }),

  rest.post(endpoints.ENDPOINT_ADD_ROLE, async (req, res, ctx) => {
    const body = req.body as NewUserRole;

    await new Promise((r) => setTimeout(r, RESPONSE_TIME));

    if (!isAuthorized([UserRole.Admin])) {
      return res(ctx.status(403), ctx.json({ message: 'You are not authorized to add a role' }));
    }

    mockUserDB.push({ id: uid(), ...body });

    return res(ctx.status(200));
  }),

  rest.post(endpoints.ENDPOINT_REMOVE_ROLE, async (req, res, ctx) => {
    const body = req.body as RemoveUserRole;
    await new Promise((r) => setTimeout(r, RESPONSE_TIME));

    if (!isAuthorized([UserRole.Admin])) {
      return res(ctx.status(403), ctx.json({ message: 'You are not authorized to remove a role' }));
    }
    mockUserDB = mockUserDB.filter((user) => user.sciper !== body.sciper);

    return res(ctx.status(200));
  }),

  rest.get(endpoints.getProxiesAddresses(':ElectionID'), async (req, res, ctx) => {
    const { ElectionID } = req.params;
    await new Promise((r) => setTimeout(r, RESPONSE_TIME));

    console.log(mockNodeProxyAddresses);
    const response = [];
    mockNodeProxyAddresses
      .get(ElectionID as string)
      .forEach((proxy, node) => response.push({ [node]: proxy }));

    return res(ctx.status(200), ctx.json({ Proxies: response }));
  }),

  rest.put(endpoints.editProxiesAddresses(':ElectionID'), async (req, res, ctx) => {
    const { ElectionID } = req.params;
    const body = req.body as NewProxyAddress;

    const newNodeProxyAddresses = new Map();

    body.Proxies.forEach((value) => {
      Object.entries(value).forEach((v) => newNodeProxyAddresses.set(v[0], v[1]));
    });

    mockNodeProxyAddresses.set(ElectionID as string, newNodeProxyAddresses);
    console.log(mockNodeProxyAddresses);

    await new Promise((r) => setTimeout(r, RESPONSE_TIME));

    return res(ctx.status(200), ctx.text('Action successfully done'));
  }),

  rest.get(endpoints.getProxyAddress(':ElectionID', '*'), async (req, res, ctx) => {
    const { ElectionID } = req.params;

    const node0 = mockRoster[0];
    const proxy = mockNodeProxyAddresses.get(ElectionID as string).get(node0);

    await new Promise((r) => setTimeout(r, RESPONSE_TIME));

    return res(ctx.status(200), ctx.json({ [node0]: proxy }));
  }),
];
