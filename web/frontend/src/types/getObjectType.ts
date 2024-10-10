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
      URL: '',
    },
    Scaffold: [],
    AdditionalInfo: '',
  };
};

const newSubject = (): types.Subject => {
  return {
    ID: uid(),
    Title: {
      En: '',
      Fr: '',
      De: '',
      URL: '',
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
      URL: '',
    },
    MaxN: 2,
    MinN: 2,
    Choices: [],
    ChoicesMap: { ChoicesMap: new Map(Object.entries(obj)), URLs: [''] },
    Type: RANK,
    Hint: {
      En: '',
      Fr: '',
      De: '',
    },
  };
};

const newSelect = (): types.SelectQuestion => {
  return {
    ID: uid(),
    Title: {
      En: '',
      Fr: '',
      De: '',
      URL: '',
    },
    MaxN: 1,
    MinN: 1,
    Choices: [],
    ChoicesMap: { ChoicesMap: new Map(Object.entries(obj)), URLs: [''] },
    Type: SELECT,
    Hint: {
      En: '',
      Fr: '',
      De: '',
    },
  };
};

const newText = (): types.TextQuestion => {
  return {
    ID: uid(),
    Title: {
      En: '',
      Fr: '',
      De: '',
      URL: '',
    },
    MaxN: 1,
    MinN: 0,
    MaxLength: 50,
    Regex: '',
    Choices: [],
    ChoicesMap: { ChoicesMap: new Map(Object.entries(obj)), URLs: [''] },
    Type: TEXT,
    Hint: {
      En: '',
      Fr: '',
      De: '',
    },
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

const choicesToChoicesMap = (choices: types.Choice[]): types.ChoicesMap => {
  const choicesMap = { ChoicesMap: new Map<string, string[]>(), URLs: [] };

  // choices is of form `{"en": "choice1", "fr": "choix1"}`
  choices.forEach((choice) => {
    const choiceObj = JSON.parse(choice.Choice) as { [key: string]: string };
    for (const [lang, c] of Object.entries(choiceObj)) {
      if (!choicesMap.ChoicesMap.has(lang)) {
        choicesMap.ChoicesMap.set(lang, [c]);
      } else {
        choicesMap.ChoicesMap.get(lang).push(c);
      }
    }
    choicesMap.URLs.push(choice.URL);
  });

  return choicesMap;
};

const choicesMapToChoices = (ChoicesMap: types.ChoicesMap): types.Choice[] => {
  let choices: types.Choice[] = [];
  for (let i = 0; i < ChoicesMap.ChoicesMap.get('en').length; i++) {
    const choiceMap = new Map<string, string>();
    for (let key of ChoicesMap.ChoicesMap.keys()) {
      if (ChoicesMap.ChoicesMap.get(key)[i] === '') {
        continue;
      }
      choiceMap.set(key, ChoicesMap.ChoicesMap.get(key)[i]);
    }
    choices.push({
      Choice: JSON.stringify(Object.fromEntries(choiceMap)),
      URL: ChoicesMap.URLs[i],
    });
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
  elements.forEach((element) => {
    switch (element.Type) {
      case RANK:
        rankQuestion.push({
          ...(element as types.RankQuestion),
          Title: element.Title,
          Choices: choicesMapToChoices(
            (element as types.RankQuestion).ChoicesMap
          ),
        });
        break;
      case SELECT:
        selectQuestion.push({
          ...(element as types.SelectQuestion),
          Title: element.Title,
          Choices: choicesMapToChoices(
            (element as types.SelectQuestion).ChoicesMap
          ),
        });
        break;
      case TEXT:
        textQuestion.push({
          ...(element as types.TextQuestion),
          Title: element.Title,
          Choices: choicesMapToChoices(
            (element as types.TextQuestion).ChoicesMap
          ),
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
