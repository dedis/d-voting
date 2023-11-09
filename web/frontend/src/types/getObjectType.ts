import ShortUniqueId from 'short-unique-id';
import * as types from './configuration';
import { ID, RANK, SELECT, SUBJECT, TEXT } from './configuration';

const uid: Function = new ShortUniqueId({ length: 8 });

const emptyConfiguration = (): types.Configuration => {
  return {
    Title: {
      En: '',
      Fr: '',
      De: '',
    },
    Scaffold: [],
  };
};

const newSubject = (): types.Subject => {
  return {
    ID: uid(),
    Title: {
      En: '',
      Fr: '',
      De: '',
    },
    Order: [],
    Type: SUBJECT,
    Elements: new Map(),
  };
};
const obj = { en: [''], fr: [''], de: [''] };
const newRank = (): types.RankQuestion => {
  return {
    ID: uid(),
    Title: {
      En: '',
      Fr: '',
      De: '',
    },
    MaxN: 2,
    MinN: 2,
    Choices: [],
    ChoicesMap: new Map(Object.entries(obj)),
    Type: RANK,
    Hint: '',
    HintFr: '',
    HintDe: '',
  };
};

const newSelect = (): types.SelectQuestion => {
  return {
    ID: uid(),
    Title: {
      En: '',
      Fr: '',
      De: '',
    },
    MaxN: 1,
    MinN: 1,
    Choices: [],
    ChoicesMap: new Map(Object.entries(obj)),
    Type: SELECT,
    Hint: '',
    HintFr: '',
    HintDe: '',
  };
};

const newText = (): types.TextQuestion => {
  return {
    ID: uid(),
    Title: {
      En: '',
      Fr: '',
      De: '',
    },
    MaxN: 1,
    MinN: 0,
    MaxLength: 50,
    Regex: '',
    Choices: [],
    ChoicesMap: new Map(Object.entries(obj)),
    Type: TEXT,
    Hint: '',
    HintFr: '',
    HintDe: '',
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

const choicesToChoicesMap = (choices: string[]): Map<string, string[]> => {
  const choicesMap = new Map<string, string[]>();

  // choices is of form `{"en": "choice1", "fr": "choix1"}`
  choices.forEach((choice) => {
    const choiceObj = JSON.parse(choice) as { [key: string]: string };
    for (const [lang, c] of Object.entries(choiceObj)) {
      if (!choicesMap.has(lang)) {
        choicesMap.set(lang, [c]);
      } else {
        choicesMap.get(lang).push(c);
      }
    }
  });

  return choicesMap;
};

const choicesMapToChoices = (ChoicesMap: Map<string, string[]>): string[] => {
  let choices: string[] = [];
  for (let i = 0; i < ChoicesMap.get('en').length; i++) {
    const choiceMap = new Map<string, string>();
    for (let key of ChoicesMap.keys()) {
      if (ChoicesMap.get(key)[i] === '') {
        continue;
      }
      choiceMap.set(key, ChoicesMap.get(key)[i]);
    }
    const s = JSON.stringify(Object.fromEntries(choiceMap));
    choices.push(s);
  }
  return choices;
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
  let hint = '';
  elements.forEach((element) => {
    switch (element.Type) {
      case RANK:
        hint = JSON.stringify({
          en: (element as types.RankQuestion).Hint,
          fr: (element as types.RankQuestion).HintFr,
          de: (element as types.RankQuestion).HintDe,
        });
        rankQuestion.push({
          ...(element as types.RankQuestion),
          Title: element.Title,
          Choices: choicesMapToChoices(
            (element as types.RankQuestion).ChoicesMap
          ),
          Hint: hint,
        });
        break;
      case SELECT:
        hint = JSON.stringify({
          en: (element as types.SelectQuestion).Hint,
          fr: (element as types.SelectQuestion).HintFr,
          de: (element as types.SelectQuestion).HintDe,
        });
        selectQuestion.push({
          ...(element as types.SelectQuestion),
          Title: element.Title,
          Choices: choicesMapToChoices(
            (element as types.SelectQuestion).ChoicesMap
          ),
          Hint: hint,
        });
        break;
      case TEXT:
        hint = JSON.stringify({
          en: (element as types.TextQuestion).Hint,
          fr: (element as types.TextQuestion).HintFr,
          de: (element as types.TextQuestion).HintDe,
        });
        textQuestion.push({
          ...(element as types.TextQuestion),
          Title: element.Title,
          Choices: choicesMapToChoices(
            (element as types.TextQuestion).ChoicesMap
          ),
          Hint: hint,
        });
        break;
      case SUBJECT:
        subjects.push({
          ...(element as types.Subject),
          Title: element.Title,
        });
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
  choicesToChoicesMap,
  choicesMapToChoices,
};
