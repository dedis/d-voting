import { Configuration, Rank, Select, Subject, Text } from 'types/configuration';

const marshallText: (text: any) => Text = (text) => {
  return {
    ...text,
    Type: 'TEXT',
  };
};

const marshallRank: (rank: any) => Rank = (rank) => {
  return {
    ...rank,
    Type: 'RANK',
  };
};

const marshallSelect: (select: any) => Select = (select) => {
  return {
    ...select,
    Type: 'SELECT',
  };
};

const marshallSubject: (subject: any) => Subject = (subject) => {
  const Subjects = subject.Subjects.length
    ? subject.Subjects.map((subj: any) => marshallSubject(subj))
    : [];
  const Ranks = subject.Ranks.length ? subject.Ranks.map((rank: any) => marshallRank(rank)) : [];
  const Selects = subject.Selects.length
    ? subject.Selects.map((rank: any) => marshallSelect(rank))
    : [];
  const Texts = subject.Texts.length ? subject.Texts.map((text: any) => marshallText(text)) : [];
  return {
    ...subject,
    Subjects,
    Ranks,
    Selects,
    Texts,
    Type: 'SUBJECT',
  };
};

const marshallConfig: (json: any) => Configuration = (json) => {
  const conf = { MainTitle: json.MainTitle, Scaffold: [] };
  for (const subject of json.Scaffold) {
    conf.Scaffold.push(marshallSubject(subject));
  }
  return conf;
};

const unmarshallText: (text: Text) => any = (text) => {
  const newText: any = { ...text };
  delete newText.Type;
  return newText;
};

const unmarshallRank: (rank: Rank) => any = (rank) => {
  const newRank: any = { ...rank };
  delete newRank.Type;
  return newRank;
};

const unmarshallSelect: (select: Select) => any = (select) => {
  const newSelect: any = { ...select };
  delete newSelect.Type;
  return newSelect;
};

const unmarshallSubject: (subject: Subject) => any = (subject) => {
  const newSubject: any = { ...subject };
  delete newSubject.Type;
  newSubject.Subjects = subject.Subjects.length
    ? subject.Subjects.map((subj) => unmarshallSubject(subj))
    : [];
  newSubject.Ranks = subject.Ranks.length ? subject.Ranks.map((rank) => unmarshallRank(rank)) : [];
  newSubject.Selects = subject.Selects.length
    ? subject.Selects.map((rank) => unmarshallSelect(rank))
    : [];
  newSubject.Texts = subject.Texts.length ? subject.Texts.map((text) => unmarshallText(text)) : [];
  return newSubject;
};

const unmarshallConfig: (configuration: Configuration) => any = (configuration) => {
  const conf = { MainTitle: configuration.MainTitle, Scaffold: [] };
  for (const subject of configuration.Scaffold) {
    conf.Scaffold.push(unmarshallSubject(subject));
  }
  return conf;
};

export { marshallConfig, unmarshallConfig };
