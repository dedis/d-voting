export type ID = string;

export const RANK = 'rank';
export const SELECT = 'select';
export const SUBJECT = 'subject';
export const TEXT = 'text';
export const ROOT_ID: ID = '0';

// Rank describes a "rank" question, which requires the user to rank choices.
export interface RankQuestion {
  ID: ID;
  Title: string;
  MaxN: number;
  MinN: number;
  Choices: Array<string>;
}

// Text describes a "text" question, which allows the user to enter free text.
export interface TextQuestion {
  ID: ID;
  Title: string;
  MaxN: number;
  MinN: number;
  MaxLength: number;
  Regex?: string;
  Choices: Array<string>;
}

// Select describes a "select" question, which requires the user to select one
// or multiple choices.
export interface SelectQuestion {
  ID: ID;
  Title: string;
  MaxN: number;
  MinN: number;
  Choices: Array<string>;
}

export interface Subject {
  ID: ID;
  Title: string;
  // Order defines the order of the different question, which all have a uniq
  // identifier. This is purely for display purpose.
  Order: Array<ID>;
  Subjects?: Array<Subject>;
  Selects?: Array<SelectQuestion>;
  Ranks?: Array<RankQuestion>;
  Texts?: Array<TextQuestion>;
}

// Configuration contains the configuration of a new poll.
export interface Configuration {
  MainTitle: string;
  Scaffold: Array<Subject>;
}

export interface Question {
  Order: number;
  ParentID: ID;
  Type: string;
  Content: RankQuestion | TextQuestion | SelectQuestion | Subject;
  render(
    question: Question,
    answers?: Answers,
    setAnswers?: React.Dispatch<React.SetStateAction<Answers>>
  ): JSX.Element;
}

// RankAnswer describes a "rank" answer.
export interface RankAnswer {
  ID: ID;
  Answers: number[];
}

// TextAnswer describes a "text" answer.
export interface TextAnswer {
  ID: ID;
  Answers: string[];
}

// SelectAnswer describes a "select" answer.
export interface SelectAnswer {
  ID: ID;
  Answers: boolean[];
}

// Error contains a message for each answer of an election.
export interface Error {
  ID: ID;
  Message: string;
}

// Answers contains all the different types of answers for an election.
export interface Answers {
  SelectAnswers: SelectAnswer[];
  RankAnswers: RankAnswer[];
  TextAnswers: TextAnswer[];
  Errors: Error[];
}
