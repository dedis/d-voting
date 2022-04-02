type ID = string;

// Text describes a "text" question, which allows the user to enter free text.
interface Text {
  ID: ID;
  Title: string;
  MaxN: number;
  MinN: number;
  MaxLength: number;
  Regex?: string;
  Choices: string[];
}

// Select describes a "select" question, which requires the user to select one
// or multiple choices.
interface Select {
  ID: ID;
  Title: string;
  MaxN: number;
  MinN: number;
  Choices: string[];
}

// Rank describes a "rank" question, which requires the user to rank choices.
interface Rank {
  ID: ID;
  Title: string;
  MaxN: number;
  MinN: number;
  Choices: string[];
}

interface Subject {
  ID: ID;
  Title: string;
  Order: ID[];
  Subjects?: Subject[];
  Selects?: Select[];
  Ranks?: Rank[];
  Texts?: Text[];
}

// Configuration contains the configuration of a new poll.
interface Configuration {
  MainTitle: string;
  Scaffold: Subject[];
}

export type { ID, Text, Select, Rank, Subject, Configuration };
