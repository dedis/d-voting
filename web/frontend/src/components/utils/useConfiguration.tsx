import { useEffect, useState } from 'react';
import { Answers, Configuration, ID, Subject } from 'types/configuration';
import { subjectFromJSON } from 'types/JSONparser';

// Take a JSON object and unmarshal it into a Configuration
// Returns a Configuration and the initialized Answers
const useConfiguration = (configObj: any) => {
  const [configuration, setConfiguration]: [
    Configuration,
    React.Dispatch<React.SetStateAction<Configuration>>
  ] = useState({ MainTitle: '', Scaffold: new Array<Subject>() });

  const [answers, setAnswers]: [Answers, React.Dispatch<React.SetStateAction<Answers>>] =
    useState(null);

  const configFromJSON = (config: Configuration, answerMap: Answers) => {
    let scaffold = new Array<Subject>();

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

      let config: Configuration = { MainTitle: '', Scaffold: new Array<Subject>() };
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
