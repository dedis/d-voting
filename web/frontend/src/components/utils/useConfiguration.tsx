import { useEffect, useState } from 'react';
import { ID, Rank, Select, Subject, Text } from './types';

export const RANK = 'rank';
export const SELECT = 'select';
export const SUBJECT = 'subject';
export const TEXT = 'text';
export const ROOT_ID: ID = '0';

export interface Question {
  Order: number;
  ParentID: ID;
  Type: string;
  Content: Select | Subject | Rank | Text;
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

export interface Answers {
  SelectAnswers: SelectAnswer[];
  RankAnswers: RankAnswer[];
  TextAnswers: TextAnswer[];
  Errors: Error[];
}

function initConfiguration(
  question: any,
  parentId: ID,
  sortedQuestions: Array<Question>,
  currentOrder: number,
  answerList
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
        initConfiguration(subject, question.ID, sortedQuestions, currentOrder + order, answerList) +
        1;
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
      answerList.SelectAnswers.push({
        ID: select.ID,
        Answers: new Array<boolean>(select.Choices.length).fill(false),
      });
      answerList.Errors.push({
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
      answerList.RankAnswers.push({
        ID: rank.ID,
        Answers: Array.from(Array(rank.Choices.length).keys()),
      });
      answerList.Errors.push({
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
      answerList.TextAnswers.push({
        ID: text.ID,
        Answers: new Array<string>(text.Choices.length).fill(''),
      });
      answerList.Errors.push({
        ID: text.ID,
        Message: '',
      });
    }
  }

  return numberOfQuestion;
}

const useConfiguration = (configuration) => {
  const [sortedQuestions, setSortedQuestions] = useState(Array<Question>());
  const [answers, setAnswers]: [Answers, React.Dispatch<React.SetStateAction<Answers>>] =
    useState(null);
  useEffect(() => {
    if (configuration !== null) {
      let order: number = 0;
      let questionList = Array<Question>();
      let answerList: Answers = {
        SelectAnswers: new Array<SelectAnswer>(),
        RankAnswers: new Array<RankAnswer>(),
        TextAnswers: new Array<TextAnswer>(),
        Errors: new Array<Error>(),
      };

      for (const subject of configuration.Scaffold) {
        order += initConfiguration(subject, ROOT_ID, questionList, order, answerList);
      }
      questionList.sort((q1, q2) => {
        return q1.Order - q2.Order;
      });
      setSortedQuestions(questionList);
      setAnswers(answerList);
    }
  }, [configuration]);

  return {
    sortedQuestions,
    answers,
    setAnswers,
  };
};

export default useConfiguration;
