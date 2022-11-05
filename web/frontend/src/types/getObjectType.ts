import ShortUniqueId from 'short-unique-id';
import * as types from './configuration';
import { ID, RANK, SELECT, SUBJECT, TEXT } from './configuration';

const uid: Function = new ShortUniqueId({ length: 8 });

const emptyConfiguration = (): types.Configuration => {
  return {
    MainTitle: '',
    Scaffold: [],
    TitleLg1: '',
    //ScaffoldLg1: [],
    TitleLg2: '',
    //ScaffoldLg2: [],
  };
};

const newSubject = (): types.Subject => {
  return {
    ID: uid(),
    Title: '',
    TitleFr: '',
    TitleDe: '',
    Order: [],
    Type: SUBJECT,
    Elements: new Map(),
  };
};

const newRank = (): types.RankQuestion => {
  return {
    ID: uid(),
    Title: '',
    TitleFr: '',
    TitleDe: '',
    MaxN: 2,
    MinN: 2,
    Choices: ['', ''],
    ChoicesDe: ['',''],
    ChoicesFr: ['',''],
    Type: RANK,
  };
};

const newSelect = (): types.SelectQuestion => {
  return {
    ID: uid(),
    Title: '',
    TitleDe: '',
    TitleFr: '',
    MaxN: 1,
    MinN: 1,
    Choices: [''],
    ChoicesDe: [''],
    ChoicesFr: [''],
    Type: SELECT,
  };
};

const newText = (): types.TextQuestion => {
  return {
    ID: uid(),
    Title: '',
    TitleFr: '',
    TitleDe: '',
    MaxN: 1,
    MinN: 0,
    MaxLength: 50,
    Regex: '',
    Choices: [''],
    ChoicesDe: [''],
    ChoicesFr: [''],
    Type: TEXT,
  };
};

const newAnswer = (): types.Answers => {
  return {
    SelectAnswers: new Map<ID, boolean[]>(),
    RankAnswers: new Map<ID, number[]>(),
    TextAnswers: new Map<ID, string[]>(),
    Errors: new Map<ID, string>(),
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
  newAnswer,
  newSubject,
  newRank,
  newSelect,
  newText,
  answersFrom,
  toArraysOfSubjectElement,
};
