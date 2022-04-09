import { useEffect, useState } from 'react';
import { Answers, Configuration } from 'types/configuration';
import { emptyConfiguration } from 'types/getObjectType';
import { unmarshalConfigAndCreateAnswers } from 'types/JSONparser';

// Take a JSON object and unmarshal it into a Configuration
// Returns a Configuration and the initialized Answers
const useConfiguration = (configObj: any) => {
  const [configuration, setConfiguration]: [
    Configuration,
    React.Dispatch<React.SetStateAction<Configuration>>
  ] = useState(emptyConfiguration());

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

export default useConfiguration;
