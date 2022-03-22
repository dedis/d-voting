type ID = string;

// Text describes a "text" question, which allows the user to enter free text.
interface Text {
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
interface Select {
  ID: ID;
  Title: string;
  MaxN: number;
  MinN: number;
  Choices: Array<string>;
}

// Rank describes a "rank" question, which requires the user to rank choices.
interface Rank {
  ID: ID;
  Title: string;
  MaxN: number;
  MinN: number;
  Choices: Array<string>;
}

interface Subject {
  ID: ID;
  Title: string;
  // Order defines the order of the different question, which all have a uniq
  // identifier. This is purely for display purpose.
  Order: Array<ID>;
  Subjects?: Array<Subject>;
  Selects?: Array<Select>;
  Ranks?: Array<Rank>;
  Texts?: Array<Text>;
}

// Configuration contains the configuration of a new poll.
interface Configuration {
  MainTitle: string;
  Scaffold: Array<Subject>;
}

export type { ID, Text, Select, Rank, Subject, Configuration };
