import ShortUniqueId from 'short-unique-id';
import { Rank, Select, Subject, Text } from '../../../../types/configuration';

const uid: Function = new ShortUniqueId({ length: 8 });

const newSubject: () => Subject = () => {
  return {
    ID: uid(),
    Title: '',
    Order: [],
    Subjects: [],
    Ranks: [],
    Selects: [],
    Texts: [],
    Type: 'SUBJECT',
  };
};

const newRank: () => Rank = () => {
  return {
    ID: uid(),
    Title: '',
    MaxN: 0,
    MinN: 0,
    Choices: [],
    Type: 'RANK',
  };
};

const newSelect: () => Select = () => {
  return {
    ID: uid(),
    Title: '',
    MaxN: 0,
    MinN: 0,
    Choices: [],
    Type: 'SELECT',
  };
};

const newText: () => Text = () => {
  return {
    ID: uid(),
    Title: '',
    MaxN: 0,
    MinN: 0,
    MaxLength: 50,
    Regex: '',
    Choices: [],
    Type: 'TEXT',
  };
};

export { newSubject, newRank, newSelect, newText };
