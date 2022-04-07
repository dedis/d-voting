export type ID = string;

export const RANK: string = 'rank';
export const SELECT: string = 'select';
export const SUBJECT: string = 'subject';
export const TEXT: string = 'text';
export const ROOT_ID: ID = '0';

export interface SubjectElement {
  ID: ID;
  Type: string;
  Title: string;
}

// Rank describes a "rank" question, which requires the user to rank choices.
export interface RankQuestion extends SubjectElement {
  MaxN: number;
  MinN: number;
  Choices: Array<string>;
}

// Text describes a "text" question, which allows the user to enter free text.
export interface TextQuestion extends SubjectElement {
  MaxN: number;
  MinN: number;
  MaxLength: number;
  Regex: string;
  Choices: Array<string>;
}

// Select describes a "select" question, which requires the user to select one
// or multiple choices.
export interface SelectQuestion extends SubjectElement {
  MaxN: number;
  MinN: number;
  Choices: Array<string>;
}

export interface Subject extends SubjectElement {
  Order: Array<ID>;
  Elements: Map<ID, SubjectElement>;
}

// Configuration contains the configuration of a new poll.
export interface Configuration {
  MainTitle: string;
  Scaffold: Array<SubjectElement>;
}

// Answers describes the current answers for each type of question
// as well as a possible Error message
export interface Answers {
  SelectAnswers: Map<ID, boolean[]>;
  RankAnswers: Map<ID, number[]>;
  TextAnswers: Map<ID, string[]>;
  Errors: Map<ID, string>;
}

// Create a deep copy of an answer
export function answersFrom(answers: Answers) {
  let newAnswers: Answers = {
    SelectAnswers: new Map(answers.SelectAnswers),
    RankAnswers: new Map(answers.RankAnswers),
    TextAnswers: new Map(answers.TextAnswers),
    Errors: new Map(answers.Errors),
  };
  return newAnswers;
}

function rankFromJSON(obj: any): RankQuestion {
  let rank: RankQuestion = {
    ID: obj.ID,
    Title: obj.Title,
    Type: RANK,
    MaxN: obj.MaxN,
    MinN: obj.MinN,
    Choices: obj.Choices,
  };
  return rank;
}

function textFromJSON(obj: any): TextQuestion {
  let text: TextQuestion = {
    ID: obj.ID,
    Title: obj.Title,
    Type: TEXT,
    MaxN: obj.MaxN,
    MinN: obj.MinN,
    Choices: obj.Choices,
    MaxLength: obj.MaxLength,
    Regex: obj.Regex,
  };
  return text;
}

function selectFromJSON(obj: any): SelectQuestion {
  let select: SelectQuestion = {
    ID: obj.ID,
    Title: obj.Title,
    Type: SELECT,
    MaxN: obj.MaxN,
    MinN: obj.MinN,
    Choices: obj.Choices,
  };
  return select;
}

// Create a subject form a JSON object and initializes the Answers at
// the same time (so as to parse only once through the whole structure)
export function subjectFromJSON(obj: any, answerMap: Answers) {
  const elements = new Map<ID, SubjectElement>();

  for (const subjectObj of obj.Subjects) {
    let subject = subjectFromJSON(subjectObj, answerMap);
    elements.set(subject.ID, subject);
  }

  for (const rankObj of obj.Ranks) {
    let rank = rankFromJSON(rankObj);
    elements.set(rank.ID, rank);
    answerMap.RankAnswers.set(rank.ID, Array.from(Array(rank.Choices.length).keys()));
    answerMap.Errors.set(rank.ID, '');
  }

  for (const textObj of obj.Texts) {
    let text = textFromJSON(textObj);
    elements.set(text.ID, text);
    answerMap.TextAnswers.set(text.ID, new Array<string>(text.Choices.length).fill(''));
    answerMap.Errors.set(text.ID, '');
  }

  for (const selectObj of obj.Selects) {
    let select = selectFromJSON(selectObj);
    elements.set(select.ID, select);
    answerMap.SelectAnswers.set(select.ID, new Array<boolean>(select.Choices.length).fill(false));
    answerMap.Errors.set(select.ID, '');
  }

  let subject: Subject = {
    ID: obj.ID,
    Title: obj.Title,
    Type: SUBJECT,
    Order: obj.Order,
    Elements: elements,
  };

  return subject;
}
