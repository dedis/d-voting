import { ID, RANK, SELECT, SUBJECT, TEXT } from 'types/configuration';
import * as types from 'types/configuration';
import { toArraysOfSubjectElement } from './getObjectType';

const unmarshalText = (text: any): types.TextQuestion => {
  return {
    ...text,
    Type: TEXT,
  };
};

const unmarshalRank = (rank: any): types.RankQuestion => {
  return {
    ...rank,
    Type: RANK,
  };
};

const unmarshalSelect = (select: any): types.SelectQuestion => {
  return {
    ...select,
    Type: SELECT,
  };
};

const unmarshalSubject = (subjectObj: any): types.Subject => {
  const elements = new Map<ID, types.SubjectElement>();

  for (const subSubjectObj of subjectObj.Subjects) {
    const subSubject = unmarshalSubject(subSubjectObj);
    elements.set(subSubject.ID, subSubject);
  }

  for (const rankObj of subjectObj.Ranks) {
    const rank = unmarshalRank(rankObj);
    elements.set(rank.ID, rank);
  }

  for (const textObj of subjectObj.Texts) {
    const text = unmarshalText(textObj);
    elements.set(text.ID, text);
  }

  for (const selectObj of subjectObj.Selects) {
    const select = unmarshalSelect(selectObj);
    elements.set(select.ID, select);
  }

  return {
    ...subjectObj,
    Type: SUBJECT,
    Elements: elements,
  };
};

// Create a subject from a JSON object and initializes the Answers at
// the same time (so as to go only once through the whole Scaffold)
const unmarshalSubjectAndCreateAnswers = (
  subjectObj: any,
  answerMap: types.Answers
): types.Subject => {
  const elements = new Map<ID, types.SubjectElement>();

  for (const subSubjectObj of subjectObj.Subjects) {
    const subSubject = unmarshalSubjectAndCreateAnswers(subSubjectObj, answerMap);
    elements.set(subSubject.ID, subSubject);
  }

  for (const rankObj of subjectObj.Ranks) {
    const rank = unmarshalRank(rankObj);
    elements.set(rank.ID, rank);
    answerMap.RankAnswers.set(rank.ID, Array.from(Array(rank.Choices.length).keys()));
    answerMap.Errors.set(rank.ID, '');
  }

  for (const selectObj of subjectObj.Selects) {
    const select = unmarshalSelect(selectObj);
    elements.set(select.ID, select);
    answerMap.SelectAnswers.set(select.ID, new Array<boolean>(select.Choices.length).fill(false));
    answerMap.Errors.set(select.ID, '');
  }

  for (const textObj of subjectObj.Texts) {
    const text = unmarshalText(textObj);
    elements.set(text.ID, text);
    answerMap.TextAnswers.set(text.ID, new Array<string>(text.Choices.length).fill(''));
    answerMap.Errors.set(text.ID, '');
  }

  return {
    ...subjectObj,
    Type: SUBJECT,
    Elements: elements,
  };
};

const unmarshalConfig = (json: any): types.Configuration => {
  const conf = { MainTitle: json.MainTitle, Scaffold: [] };
  for (const subject of json.Scaffold) {
    conf.Scaffold.push(unmarshalSubject(subject));
  }
  return conf;
};

const unmarshalConfigAndCreateAnswers = (
  configObj: any
): { newConfiguration: types.Configuration; newAnswers: types.Answers } => {
  const scaffold = new Array<types.Subject>();
  const newAnswers: types.Answers = {
    SelectAnswers: new Map<ID, boolean[]>(),
    RankAnswers: new Map<ID, number[]>(),
    TextAnswers: new Map<ID, string[]>(),
    Errors: new Map<ID, string>(),
  };

  for (const subjectObj of configObj.Scaffold) {
    let subject = unmarshalSubjectAndCreateAnswers(subjectObj, newAnswers);
    scaffold.push(subject);
  }

  const newConfiguration = { ...configObj, Scaffold: scaffold };

  return { newConfiguration, newAnswers };
};

const marshalText = (text: types.TextQuestion): any => {
  const newText: any = { ...text };
  delete newText.Type;
  return newText;
};

const marshalRank = (rank: types.RankQuestion): any => {
  const newRank: any = { ...rank };
  delete newRank.Type;
  return newRank;
};

const marshalSelect = (select: types.SelectQuestion): any => {
  const newSelect: any = { ...select };
  delete newSelect.Type;
  return newSelect;
};

const marshalSubject = (subject: types.Subject): any => {
  const newSubject: any = { ...subject };
  const { rankQuestion, selectQuestion, textQuestion, subjects } = toArraysOfSubjectElement(
    subject.Elements
  );
  delete newSubject.Type;
  delete newSubject.Elements;

  newSubject.Ranks = new Array<any>();
  newSubject.Selects = new Array<any>();
  newSubject.Texts = new Array<any>();
  newSubject.Subjects = new Array<any>();

  rankQuestion.forEach((rank) => newSubject.Ranks.push(marshalRank(rank)));
  selectQuestion.forEach((select) => newSubject.Selects.push(marshalSelect(select)));
  textQuestion.forEach((text) => newSubject.Texts.push(marshalText(text)));
  subjects.forEach((subj) => newSubject.Subjects.push(marshalSubject(subj)));

  return newSubject;
};

const marshalConfig = (configuration: types.Configuration): any => {
  const conf = { MainTitle: configuration.MainTitle, Scaffold: [] };
  for (const subject of configuration.Scaffold) {
    conf.Scaffold.push(marshalSubject(subject));
  }
  return conf;
};

export { marshalConfig, unmarshalConfig, unmarshalConfigAndCreateAnswers };
