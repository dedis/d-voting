import { useEffect, useState } from 'react';
import { Answers, Configuration } from 'types/configuration';
import { emptyConfiguration, newAnswer } from 'types/getObjectType';
import {
  unmarshalConfig,
  unmarshalConfigAndCreateAnswers,
  unmarshalSubjectAndCreateAnswers,
} from 'types/JSONparser';

// Returns a Configuration and the initialized Answers
const useConfiguration = (configObj: any) => {
  const [configuration, setConfiguration] = useState<Configuration>(emptyConfiguration());
  const [answers, setAnswers] = useState<Answers>(null);

  useEffect(() => {
    console.log(configObj);
    if (configObj !== null) {
      const { newConfiguration, newAnswers } = unmarshalConfigAndCreateAnswers(configObj);
      setConfiguration(newConfiguration);
      setAnswers(newAnswers);
    }
  }, [configObj]);

  return {
    configuration,
    answers,
    setAnswers,
  };
};

const useConfigurationOnly = (configObj: any) => {
  const [configuration, setConfiguration] = useState<Configuration>(emptyConfiguration());

  useEffect(() => {
    if (configObj !== null) {
      const conf = unmarshalConfig(configObj);
      setConfiguration(conf);
    }
  }, [configObj]);

  return configuration;
};

// Custom hook to create answers from a Configuration.
const useAnswers = (configuration: Configuration) => {
  const [answers, setAnswers] = useState<Answers>(null);

  useEffect(() => {
    if (configuration !== null) {
      const newAnswers: Answers = newAnswer();

      for (const subjectObj of configuration.Scaffold) {
        unmarshalSubjectAndCreateAnswers(subjectObj, newAnswers);
      }

      setAnswers(newAnswers);
    }
  }, [configuration]);

  return { answers, setAnswers };
};

export { useConfiguration, useConfigurationOnly, useAnswers };
