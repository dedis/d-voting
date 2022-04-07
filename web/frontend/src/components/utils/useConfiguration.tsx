import { useEffect, useState } from 'react';
import { Answers, Configuration, ID, SubjectElement, subjectFromJSON } from 'types/configuration';

// Take a JSON object and unmarshal it into a Configuration
// Returns a Configuration and the initialized Answers
const useConfiguration = (configObj: any) => {
  const [configuration, setConfiguration]: [
    Configuration,
    React.Dispatch<React.SetStateAction<Configuration>>
  ] = useState({ MainTitle: '', Scaffold: new Array<SubjectElement>() });

  const [answers, setAnswers]: [Answers, React.Dispatch<React.SetStateAction<Answers>>] =
    useState(null);

  const configFromJSON = (config: Configuration, answerMap: Answers) => {
    let scaffold = new Array<SubjectElement>();

    for (const subjectObj of configObj.Scaffold) {
      let subject = subjectFromJSON(subjectObj, answerMap);
      scaffold.push(subject);
    }

    config.MainTitle = configObj.MainTitle;
    config.Scaffold = scaffold;
  };

  useEffect(() => {
    if (configObj !== null) {
      let answerMap: Answers = {
        SelectAnswers: new Map<ID, boolean[]>(),
        RankAnswers: new Map<ID, number[]>(),
        TextAnswers: new Map<ID, string[]>(),
        Errors: new Map<ID, string>(),
      };

      let config: Configuration = { MainTitle: '', Scaffold: new Array<SubjectElement>() };
      configFromJSON(config, answerMap);
      setConfiguration(config);
      setAnswers(answerMap);
    }
  }, [configObj]);

  return {
    configuration,
    answers,
    setAnswers,
  };
};

export default useConfiguration;
