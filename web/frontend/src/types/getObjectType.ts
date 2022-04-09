import ShortUniqueId from 'short-unique-id';
import {
  Answers,
  RANK,
  RankQuestion,
  SELECT,
  SUBJECT,
  SelectQuestion,
  Subject,
  TEXT,
  TextQuestion,
} from './configuration';

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
    Type: SUBJECT,
    Elements: new Map(),
  };
};

const newRank: () => RankQuestion = () => {
  return {
    ID: uid(),
    Title: '',
    MaxN: 0,
    MinN: 0,
    Choices: [],
    Type: RANK,
  };
};

const newSelect: () => SelectQuestion = () => {
  return {
    ID: uid(),
    Title: '',
    MaxN: 0,
    MinN: 0,
    Choices: [],
    Type: SELECT,
  };
};

const newText: () => TextQuestion = () => {
  return {
    ID: uid(),
    Title: '',
    MaxN: 0,
    MinN: 0,
    MaxLength: 50,
    Regex: '',
    Choices: [],
    Type: TEXT,
  };
};

// Create a deep copy of the answers
const answersFrom = (answers: Answers) => {
  let newAnswers: Answers = {
    SelectAnswers: new Map(answers.SelectAnswers),
    RankAnswers: new Map(answers.RankAnswers),
    TextAnswers: new Map(answers.TextAnswers),
    Errors: new Map(answers.Errors),
  };
  return newAnswers;
};

export { newSubject, newRank, newSelect, newText, answersFrom };
