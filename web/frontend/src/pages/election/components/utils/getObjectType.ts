import ShortUniqueId from 'short-unique-id';
import { ID, Rank, Select, Subject, Text } from '../../../../types/configuration';

const uid: Function = new ShortUniqueId({ length: 8 });

const newSubject: (SubjectID?: ID) => Subject = (SubjectID) => {
  return {
    ID: SubjectID ? SubjectID : uid(),
    Title: '',
    Order: [],
    Subjects: [],
    Ranks: [],
    Selects: [],
    Texts: [],
  };
};

const newRank: (RankID?: ID) => Rank = (RankID) => {
  return {
    ID: RankID ? RankID : uid(),
    Title: '',
    MaxN: 0,
    MinN: 0,
    Choices: [],
  };
};

const newSelect: (SelectID?: ID) => Select = (SelectID) => {
  return {
    ID: SelectID ? SelectID : uid(),
    Title: '',
    MaxN: 0,
    MinN: 0,
    Choices: [],
  };
};

const newText: (TextID?: ID) => Text = (TextID) => {
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

export { newSubject, newRank, newSelect, newText };
