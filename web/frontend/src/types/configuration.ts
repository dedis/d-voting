type ID = string;

interface Question {
  ID: ID;
  Title: string;
  MaxN: number;
  MinN: number;
  Choices: string[];
  Type: string;
}

// Text describes a "text" question, which allows the user to enter free text.
interface Text extends Question {
  Type: 'TEXT';
  MaxLength: number;
  Regex?: string;
}

// Select describes a "select" question, which requires the user to select one
interface Select extends Question {
  Type: 'SELECT';
}

// Rank describes a "rank" question, which requires the user to rank choices.
interface Rank extends Question {
  Type: 'RANK';
}

interface Subject {
  ID: ID;
  Title: string;
  Order: ID[];
  Subjects: Subject[];
  Selects: Select[];
  Ranks: Rank[];
  Texts: Text[];
  Type: 'SUBJECT';
}

// Configuration contains the configuration of a new poll.
interface Configuration {
  MainTitle: string;
  Scaffold: Subject[];
}

export type { ID, Text, Select, Rank, Subject, Configuration };
