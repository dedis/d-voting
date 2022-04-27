import { useEffect, useState } from 'react';
import { Answers, Configuration } from 'types/configuration';
import { emptyConfiguration } from 'types/getObjectType';
import { unmarshalConfig, unmarshalConfigAndCreateAnswers } from 'types/JSONparser';

// Returns a Configuration and the initialized Answers
const useConfiguration = (configObj: any) => {
  const [configuration, setConfiguration] = useState<Configuration>(emptyConfiguration());

  const [answers, setAnswers]: [Answers, React.Dispatch<React.SetStateAction<Answers>>] =
    useState(null);

  useEffect(() => {
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

export { useConfiguration, useConfigurationOnly };
