import ShortUniqueId from 'short-unique-id';
import {
  Answers,
  Configuration,
  ID,
  RANK,
  RankQuestion,
  SELECT,
  SUBJECT,
  SelectQuestion,
  Subject,
  SubjectElement,
  TEXT,
  TextQuestion,
} from './configuration';

const uid: Function = new ShortUniqueId({ length: 8 });

const emptyConfiguration = (): Configuration => {
  return { MainTitle: '', Scaffold: [] };
};

const newSubject = (): Subject => {
  return {
    ID: uid(),
    Title: '',
    Order: [],
    Type: SUBJECT,
    Elements: new Map(),
  };
};

const newRank = (): RankQuestion => {
  return {
    ID: uid(),
    Title: '',
    MaxN: 0,
    MinN: 0,
    Choices: [],
    Type: RANK,
  };
};

const newSelect = (): SelectQuestion => {
  return {
    ID: uid(),
    Title: '',
    MaxN: 0,
    MinN: 0,
    Choices: [],
    Type: SELECT,
  };
};

const newText = (): TextQuestion => {
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
const answersFrom = (answers: Answers): Answers => {
  return {
    SelectAnswers: new Map(answers.SelectAnswers),
    RankAnswers: new Map(answers.RankAnswers),
    TextAnswers: new Map(answers.TextAnswers),
    Errors: new Map(answers.Errors),
  };
};

const toArraysOfSubjectElement = (
  elements: Map<ID, SubjectElement>
): {
  rankQuestion: RankQuestion[];
  selectQuestion: SelectQuestion[];
  textQuestion: TextQuestion[];
  subjects: Subject[];
} => {
  let rankQuestion: RankQuestion[] = new Array<RankQuestion>();
  let selectQuestion: SelectQuestion[] = new Array<SelectQuestion>();
  let textQuestion: TextQuestion[] = new Array<TextQuestion>();
  let subjects: Subject[] = new Array<Subject>();

  elements.forEach((element) => {
    switch (element.Type) {
      case RANK:
        rankQuestion.push(element as RankQuestion);
        break;
      case SELECT:
        selectQuestion.push(element as SelectQuestion);
        break;
      case TEXT:
        textQuestion.push(element as TextQuestion);
        break;
      case SUBJECT:
        subjects.push(element as Subject);
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
