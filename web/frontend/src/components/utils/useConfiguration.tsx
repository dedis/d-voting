import { renderRank, renderSelect, renderSubject, renderText } from 'pages/ballot/QuestionDisplay';
import { useEffect, useState } from 'react';
import {
  Answers,
  Configuration,
  Error,
  ID,
  Question,
  RANK,
  ROOT_ID,
  RankAnswer,
  SELECT,
  SUBJECT,
  SelectAnswer,
  Subject,
  TEXT,
  TextAnswer,
} from 'types/configuration';

// Flattens and re-orders the questions (as the first object encountered in a
// Subject might not be the first that needs to be displayed according to the Order).
// Adds a ParentID for later rendering and initializes the Answers.
function flattenSubject(
  subject: Subject,
  parentId: ID,
  sortedQuestions: Array<Question>,
  outerOrder: number,
  answerList: Answers
) {
  sortedQuestions.push({
    Order: outerOrder,
    ParentID: parentId,
    Type: SUBJECT,
    Content: subject,
    render: renderSubject,
  });

  // The number of Question inside of a top-level Subjects (i.e a with ParentID = ROOT_ID)
  let numberOfQuestion = 1;

  if (subject.Subjects) {
    for (const subSubject of subject.Subjects) {
      let insideOrder = subject.Order.findIndex((id: ID) => id === subSubject.ID) + 1;

      numberOfQuestion +=
        flattenSubject(
          subSubject,
          subject.ID,
          sortedQuestions,
          outerOrder + insideOrder,
          answerList
        ) + 1;
    }
  }

  if (subject.Selects) {
    for (const select of subject.Selects) {
      let insideOrder = subject.Order.findIndex((id: ID) => id === select.ID) + 1;
      numberOfQuestion += 1;

      sortedQuestions.push({
        Order: outerOrder + insideOrder,
        ParentID: subject.ID,
        Type: SELECT,
        Content: select,
        render: renderSelect,
      });

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

  if (subject.Ranks) {
    for (const rank of subject.Ranks) {
      let insideOrder = subject.Order.findIndex((id: ID) => id === rank.ID) + 1;
      numberOfQuestion += 1;

      sortedQuestions.push({
        Order: outerOrder + insideOrder,
        ParentID: subject.ID,
        Type: RANK,
        Content: rank,
        render: renderRank,
      });

      answerList.RankAnswers.push({
        ID: rank.ID,
        Answers: Array.from(Array(rank.Choices.length).keys()),
      });
    }
  }

  if (subject.Texts) {
    for (const text of subject.Texts) {
      let insideOrder = subject.Order.findIndex((id: ID) => id === text.ID) + 1;
      numberOfQuestion += 1;

      sortedQuestions.push({
        Order: outerOrder + insideOrder,
        ParentID: subject.ID,
        Type: TEXT,
        Content: text,
        render: renderText,
      });

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

// On Configuration changes, iterate over the top-level Subjects of the Configuration
// to sort the questions and initialize the corresponding Answers.
const useConfiguration = (configuration: Configuration) => {
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
        order += flattenSubject(subject, ROOT_ID, questionList, order, answerList);
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
