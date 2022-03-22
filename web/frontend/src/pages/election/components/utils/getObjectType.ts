import ShortUniqueId from 'short-unique-id';
import { ID, Rank, Select, Subject, Text } from '../../../../components/utils/types';

const uid: Function = new ShortUniqueId({ length: 8 });

const getObjSubject: () => Subject = () => {
  return {
    ID: uid(),
    Title: '',
    Order: [],
    Subjects: [],
    Ranks: [],
    Selects: [],
    Texts: [],
  };
};

const getObjRank: (RankID?: ID) => Rank = (RankID) => {
  return {
    ID: RankID ? RankID : uid(),
    Title: '',
    MaxN: 0,
    MinN: 0,
    Choices: [],
  };
};

const getObjSelect: (SelectID?: ID) => Select = (SelectID) => {
  return {
    ID: SelectID ? SelectID : uid(),
    Title: '',
    MaxN: 0,
    MinN: 0,
    Choices: [],
  };
};

const getObjText: (TextID?: ID) => Text = (TextID) => {
  return {
    ID: TextID ? TextID : uid(),
    Title: '',
    MaxN: 0,
    MinN: 0,
    MaxLength: 50,
    Regex: '',
    Choices: [],
  };
};

export { getObjSubject, getObjRank, getObjSelect, getObjText };
