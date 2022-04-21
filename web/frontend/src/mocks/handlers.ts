import { rest } from 'msw';
import { unmarshalConfig } from 'types/JSONparser';
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
} from '../types/frontendRequestBody';
import {
  mockElection1,
  mockElection2,
  mockElectionResult21,
  mockElectionResult22,
  mockElectionResult23,
} from './mockData';
import { ID } from 'types/configuration';
import { ElectionInfo, LightElectionInfo, STATUS } from 'types/electionInfo';

const uid = new ShortUniqueId({ length: 8 });
const mockUserID = 561934;

const mockElections: Map<ID, ElectionInfo> = new Map();
const electionID1 = uid();
const electionID2 = uid();

mockElections.set(electionID1, {
  ElectionID: electionID1,
  Status: STATUS.OPEN,
  Pubkey: 'XL4V6EMIICW',
  Result: [],
  Configuration: unmarshalConfig(mockElection1),
  BallotSize: 174,
  ChunksPerBallot: 6,
});
mockElections.set(electionID2, {
  ElectionID: electionID2,
  Status: STATUS.RESULT_AVAILABLE,
  Pubkey: 'XL4V6EMIICW',
  Result: [mockElectionResult21, mockElectionResult22, mockElectionResult23],
  Configuration: unmarshalConfig(mockElection2),
  BallotSize: 174,
  ChunksPerBallot: 6,
});

const toLightElectionInfo = (electionID: ID): LightElectionInfo => {
  const election = mockElections.get(electionID);
  console.log(election.Status);
  return {
    ElectionID: electionID,
    Title: election.Configuration.MainTitle,
    Status: election.Status,
    Pubkey: election.Pubkey,
  };
};

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
          toLightElectionInfo(election.ElectionID)
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
        Status: STATUS.OPEN,
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
    const body: EditElectionBody = JSON.parse(req.body.toString());
    const { ElectionID } = req.params;
    var Status = STATUS.INITIAL;

    switch (body.Action) {
      case 'open':
        Status = STATUS.OPEN;
        break;
      case 'close':
        Status = STATUS.CLOSED;
        break;
      case 'combineShares':
        Status = STATUS.DECRYPTED_BALLOTS;
        break;
      case 'cancel':
        Status = STATUS.CANCELED;
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
      Status: STATUS.SHUFFLED_BALLOTS,
    });

    return res(ctx.status(200), ctx.text('Action successfully done'));
  }),

  rest.put(endpoints.editDKGActors(':ElectionID'), (req, res, ctx) => {
    const { ElectionID } = req.params;
    mockElections.set(ElectionID as string, {
      ...mockElections.get(ElectionID as string),
      Status: STATUS.RESULT_AVAILABLE,
    });

    return res(ctx.status(200), ctx.text('Action successfully done'));
  }),
];
