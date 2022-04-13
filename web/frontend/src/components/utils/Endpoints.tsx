// information accessed through the middleware
export const ENDPOINT_GET_TEQ_KEY = '/api/get_teq_key';
export const ENDPOINT_PERSONAL_INFO = '/api/personal_info';
export const ENDPOINT_LOGOUT = '/api/logout';
export const ENDPOINT_USER_RIGHTS = '/api/user_rights';
export const ENDPOINT_ADD_ROLE = '/api/add_role';
export const ENDPOINT_REMOVE_ROLE = '/api/remove_role';

export const ENDPOINT_EVOTING_CREATE = '/api/evoting/elections';

export const ENDPOINT_EVOTING_ELECTION = (ElectionID = undefined) =>
  !ElectionID ? `/api/evoting/elections/:ElectionID` : `/api/evoting/elections/${ElectionID}`;

export const ENDPOINT_EVOTING_CAST_BALLOT = (ElectionID = undefined) =>
  !ElectionID
    ? `/api/evoting/elections/:ElectionID/vote`
    : `/api/evoting/elections/${ElectionID}/vote`;

export const ENDPOINT_EVOTING_SHUFFLE = (ElectionID = undefined) =>
  !ElectionID ? `/evoting/services/shuffle/:ElectionID` : `/evoting/services/shuffle/${ElectionID}`;

export const ENDPOINT_EVOTING_DECRYPT = (ElectionID = undefined) =>
  !ElectionID
    ? `/evoting/services/dkg/actors/:ElectionID`
    : `/evoting/services/dkg/actors/${ElectionID}`;

// public information can be directly fetched from dela nodes
export const ENDPOINT_EVOTING_GET_ALL = '/evoting/elections';
export const ENDPOINT_EVOTING_GET_ELECTION = (ElectionID = undefined) =>
  !ElectionID ? `/evoting/elections/:ElectionID` : `/evoting/elections/${ElectionID}`;

// To remove
export const ENDPOINT_EVOTING_RESULT = '/api/evoting/result';
