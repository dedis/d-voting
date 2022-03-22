import { useEffect, useState } from 'react';
import { ID } from './types';

export const RANK = 'rank';
export const SELECT = 'select';
export const SUBJECT = 'subject';
export const TEXT = 'text';
export const ROOT_ID: ID = '0';

export interface Question {
  Order: number;
  ParentID: ID;
  Type: string;
  Content: any;
}

export interface SelectAnswer {
  ID: ID;
  Answers: boolean[];
}

export interface RankAnswer {
  ID: ID;
  Answers: number[];
}

export interface TextAnswer {
  ID: ID;
  Answers: string[];
}

export interface Error {
  ID: ID;
  Message: string;
}

function initConfiguration(
  question: any,
  parentId: ID,
  sortedQuestions: Array<Question>,
  currentOrder: number,
  selectList: SelectAnswer[],
  rankList: RankAnswer[],
  textList: TextAnswer[],
  errors: Error[]
) {
  sortedQuestions.push({
    Order: currentOrder,
    ParentID: parentId,
    Type: SUBJECT,
    Content: question,
  });

  let numberOfQuestion = 1;
  if (question.Subjects) {
    for (const subject of question.Subjects) {
      let order = question.Order.findIndex((id: ID) => id === subject.ID) + 1;
      numberOfQuestion +=
        initConfiguration(
          subject,
          question.ID,
          sortedQuestions,
          currentOrder + order,
          selectList,
          rankList,
          textList,
          errors
        ) + 1;
    }
  }
  if (question.Selects) {
    for (const select of question.Selects) {
      let order = question.Order.findIndex((id: ID) => id === select.ID) + 1;
      sortedQuestions.push({
        Order: currentOrder + order,
        ParentID: question.ID,
        Type: SELECT,
        Content: select,
      });
      numberOfQuestion += 1;
      selectList.push({
        ID: select.ID,
        Answers: new Array<boolean>(select.Choices.length).fill(false),
      });
      errors.push({
        ID: select.ID,
        Message: '',
      });
    }
  }
  if (question.Ranks) {
    for (const rank of question.Ranks) {
      let order = question.Order.findIndex((id: ID) => id === rank.ID) + 1;
      sortedQuestions.push({
        Order: currentOrder + order,
        ParentID: question.ID,
        Type: RANK,
        Content: rank,
      });
      numberOfQuestion += 1;
      rankList.push({
        ID: rank.ID,
        Answers: Array.from(Array(rank.Choices.length).keys()),
      });
      errors.push({
        ID: rank.ID,
        Message: '',
      });
    }
  }
  if (question.Texts) {
    for (const text of question.Texts) {
      let order = question.Order.findIndex((id: ID) => id === text.ID) + 1;
      sortedQuestions.push({
        Order: currentOrder + order,
        ParentID: question.ID,
        Type: TEXT,
        Content: text,
      });
      numberOfQuestion += 1;
      textList.push({
        ID: text.ID,
        Answers: new Array<string>(text.Choices.length).fill(''),
      });
      errors.push({
        ID: text.ID,
        Message: '',
      });
    }
  }

  return numberOfQuestion;
}

const useConfiguration = (configuration) => {
  const [sortedQuestions, setSortedQuestions] = useState(Array<Question>());
  const [selectStates, setSelectStates] = useState(Array<SelectAnswer>());
  const [rankStates, setRankStates] = useState(Array<RankAnswer>());
  const [textStates, setTextStates] = useState(Array<TextAnswer>());
  const [answerErrors, setAnswerErrors] = useState(Array<Error>());
  useEffect(() => {
    if (configuration !== null) {
      let order: number = 0;
      let questionList = Array<Question>();
      let selectList = Array<SelectAnswer>();
      let rankList = Array<RankAnswer>();
      let textList = Array<TextAnswer>();
      let errors = Array<Error>();

      for (const subject of configuration.Scaffold) {
        order += initConfiguration(
          subject,
          ROOT_ID,
          questionList,
          order,
          selectList,
          rankList,
          textList,
          errors
        );
      }
      questionList.sort((q1, q2) => {
        return q1.Order - q2.Order;
      });
      setSortedQuestions(questionList);
      setSelectStates(selectList);
      setRankStates(rankList);
      setTextStates(textList);
      setAnswerErrors(errors);
    }
  }, [configuration]);

  return {
    sortedQuestions,
    selectStates,
    setSelectStates,
    rankStates,
    setRankStates,
    textStates,
    setTextStates,
    answerErrors,
    setAnswerErrors,
  };
};

export default useConfiguration;
