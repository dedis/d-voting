import ShortUniqueId from 'short-unique-id';
import * as types from './configuration';
import { ID, RANK, SELECT, SUBJECT, TEXT } from './configuration';

const uid: Function = new ShortUniqueId({ length: 8 });

const emptyConfiguration = (): types.Configuration => {
  return { MainTitle: '', Scaffold: [] };
};

const newSubject = (): types.Subject => {
  return {
    ID: uid(),
    Title: '',
    Order: [],
    Type: SUBJECT,
    Elements: new Map(),
  };
};

const newRank = (): types.RankQuestion => {
  return {
    ID: uid(),
    Title: '',
    MaxN: 2,
    MinN: 2,
    Choices: [''],
    Type: RANK,
  };
};

const newSelect = (): types.SelectQuestion => {
  return {
    ID: uid(),
    Title: '',
    MaxN: 0,
    MinN: 0,
    Choices: [''],
    Type: SELECT,
  };
};

const newText = (): types.TextQuestion => {
  return {
    ID: uid(),
    Title: '',
    MaxN: 0,
    MinN: 0,
    MaxLength: 50,
    Regex: '',
    Choices: [''],
    Type: TEXT,
  };
};

// Create a deep copy of the answers
const answersFrom = (answers: types.Answers): types.Answers => {
  return {
    SelectAnswers: new Map(answers.SelectAnswers),
    RankAnswers: new Map(answers.RankAnswers),
    TextAnswers: new Map(answers.TextAnswers),
    Errors: new Map(answers.Errors),
  };
};

const toArraysOfSubjectElement = (
  elements: Map<ID, types.SubjectElement>
): {
  rankQuestion: types.RankQuestion[];
  selectQuestion: types.SelectQuestion[];
  textQuestion: types.TextQuestion[];
  subjects: types.Subject[];
} => {
  const rankQuestion: types.RankQuestion[] = [];
  const selectQuestion: types.SelectQuestion[] = [];
  const textQuestion: types.TextQuestion[] = [];
  const subjects: types.Subject[] = [];

  elements.forEach((element) => {
    switch (element.Type) {
      case RANK:
        rankQuestion.push(element as types.RankQuestion);
        break;
      case SELECT:
        selectQuestion.push(element as types.SelectQuestion);
        break;
      case TEXT:
        textQuestion.push(element as types.TextQuestion);
        break;
      case SUBJECT:
        subjects.push(element as types.Subject);
        break;
    }
  });

  return { rankQuestion, selectQuestion, textQuestion, subjects };
};

export {
  emptyConfiguration,
  newSubject,
  newRank,
  newSelect,
  newText,
  answersFrom,
  toArraysOfSubjectElement,
};
