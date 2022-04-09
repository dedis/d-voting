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
} from 'types/configuration';

const unmarshalText: (text: any) => TextQuestion = (text) => {
  return {
    ...text,
    Type: TEXT,
  };
};

const unmarshalRank: (rank: any) => RankQuestion = (rank) => {
  return {
    ...rank,
    Type: RANK,
  };
};

const unmarshalSelect: (select: any) => SelectQuestion = (select) => {
  return {
    ...select,
    Type: SELECT,
  };
};

const unmarshalSubject: (subject: any) => Subject = (subject) => {
  const Subjects = subject.Subjects.length
    ? subject.Subjects.map((subj: any) => unmarshalSubject(subj))
    : [];
  const Ranks = subject.Ranks.length ? subject.Ranks.map((rank: any) => unmarshalRank(rank)) : [];
  const Selects = subject.Selects.length
    ? subject.Selects.map((rank: any) => unmarshalSelect(rank))
    : [];
  const Texts = subject.Texts.length ? subject.Texts.map((text: any) => unmarshalText(text)) : [];
  return {
    ...subject,
    Subjects,
    Ranks,
    Selects,
    Texts,
    Type: SUBJECT,
    Elements: new Map(),
  };
};

const unmarshalConfig: (json: any) => Configuration = (json) => {
  const conf = { MainTitle: json.MainTitle, Scaffold: [] };
  for (const subject of json.Scaffold) {
    conf.Scaffold.push(unmarshalSubject(subject));
  }
  return conf;
};

const marshalText: (text: TextQuestion) => any = (text) => {
  const newText: any = { ...text };
  delete newText.Type;
  return newText;
};

const marshalRank: (rank: RankQuestion) => any = (rank) => {
  const newRank: any = { ...rank };
  delete newRank.Type;
  return newRank;
};

const marshalSelect: (select: SelectQuestion) => any = (select) => {
  const newSelect: any = { ...select };
  delete newSelect.Type;
  return newSelect;
};

const marshalSubject: (subject: Subject) => any = (subject) => {
  const newSubject: any = { ...subject };
  delete newSubject.Type;
  newSubject.Subjects = subject.Subjects.length
    ? subject.Subjects.map((subj) => marshalSubject(subj))
    : [];
  newSubject.Ranks = subject.Ranks.length ? subject.Ranks.map((rank) => marshalRank(rank)) : [];
  newSubject.Selects = subject.Selects.length
    ? subject.Selects.map((rank) => marshalSelect(rank))
    : [];
  newSubject.Texts = subject.Texts.length ? subject.Texts.map((text) => marshalText(text)) : [];
  return newSubject;
};

const marshalConfig: (configuration: Configuration) => any = (configuration) => {
  const conf = { MainTitle: configuration.MainTitle, Scaffold: [] };
  for (const subject of configuration.Scaffold) {
    conf.Scaffold.push(marshalSubject(subject));
  }
  return conf;
};

// Create a subject form a JSON object and initializes the Answers at
// the same time (so as to parse only once through the whole structure)
export function subjectFromJSON(obj: any, answerMap: Answers): Subject {
  const elements = new Map<ID, SubjectElement>();

  for (const subjectObj of obj.Subjects) {
    let subject = subjectFromJSON(subjectObj, answerMap);
    elements.set(subject.ID, subject);
  }

  for (const rankObj of obj.Ranks) {
    let rank = unmarshalRank(rankObj);
    elements.set(rank.ID, rank);
    answerMap.RankAnswers.set(rank.ID, Array.from(Array(rank.Choices.length).keys()));
    answerMap.Errors.set(rank.ID, '');
  }

  for (const textObj of obj.Texts) {
    let text = unmarshalText(textObj);
    elements.set(text.ID, text);
    answerMap.TextAnswers.set(text.ID, new Array<string>(text.Choices.length).fill(''));
    answerMap.Errors.set(text.ID, '');
  }

  for (const selectObj of obj.Selects) {
    let select = unmarshalSelect(selectObj);
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
    Subjects: [],
    Ranks: [],
    Selects: [],
    Texts: [],
  };

  return subject;
}

export { marshalConfig, unmarshalConfig };
