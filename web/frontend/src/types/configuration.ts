type ID = string;

export const RANK: string = 'rank';
export const SELECT: string = 'select';
export const SUBJECT: string = 'subject';
export const TEXT: string = 'text';

// Title
interface Title {
  En: string;
  Fr: string;
  De: string;
}

// Hint
interface Hint {
  En: string;
  Fr: string;
  De: string;
}

// Choices
interface Choice {
  Choice: string;
  URL: string;
}
interface ChoicesMap {
  ChoicesMap: Map<string, string[]>;
  URLs: string[];
}

interface SubjectElement {
  ID: ID;
  Type: string;
  Title: Title;
}

// Rank describes a "rank" question, which requires the user to rank choices.
interface RankQuestion extends SubjectElement {
  MaxN: number;
  MinN: number;
  Choices: Choice[];
  ChoicesMap: ChoicesMap;
  Hint: Hint;
}
// Text describes a "text" question, which allows the user to enter free text.
interface TextQuestion extends SubjectElement {
  MaxN: number;
  MinN: number;
  MaxLength: number;
  Regex: string;
  Choices: Choice[];
  ChoicesMap: ChoicesMap;
  Hint: Hint;
}

// Select describes a "select" question, which requires the user to select one
// or multiple choices.
interface SelectQuestion extends SubjectElement {
  MaxN: number;
  MinN: number;
  Choices: Choice[];
  ChoicesMap: ChoicesMap;
  Hint: Hint;
}

interface Subject extends SubjectElement {
  Order: Array<ID>;
  Elements: Map<ID, SubjectElement>;
}

// Configuration contains the configuration of a new poll.
interface Configuration {
  Title: Title;
  Scaffold: Subject[];
}

// Answers describes the current answers for each type of question
// as well as a possible Error message
interface Answers {
  SelectAnswers: Map<ID, boolean[]>;
  RankAnswers: Map<ID, number[]>;
  TextAnswers: Map<ID, string[]>;
  Errors: Map<ID, string>;
}

export type {
  ID,
  Title,
  Hint,
  Choice,
  ChoicesMap,
  TextQuestion,
  SelectQuestion,
  RankQuestion,
  Subject,
  SubjectElement,
  Configuration,
  Answers,
};
