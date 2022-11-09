import { FC } from 'react';
import PropTypes from 'prop-types';

import { useTranslation } from 'react-i18next';

import { useParams } from 'react-router-dom';

import useForm from 'components/utils/useForm';
import { useConfigurationOnly } from 'components/utils/useConfiguration';

import Loading from 'pages/Loading';
import ResultExplanation from './components/ResultExplanation';
import { Tab } from '@headlessui/react';
import GroupedResult from './GroupedResult';

// Functional component that displays the result of the votes
const FormResult: FC = () => {
  const { t } = useTranslation();
  const { formId } = useParams();

  const { loading, result, configObj } = useForm(formId);
  const configuration = useConfigurationOnly(configObj);
  return (
    <div className="w-[60rem] font-sans px-4 pt-8 pb-4">
      {!loading ? (
        <div>
          <div className="flex items-center">
            <h1 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
              {t('navBarResult')}
            </h1>
            <div className="pt-1.5">
              <ResultExplanation />
            </div>
          </div>
          <div className="w-full pb-4 my-0 sm:my-4">
            <h2 className="text-lg mt-2 sm:mt-4 sm:mb-6 mb-4">
              {t('totalNumberOfVotes', { votes: result.length })}
            </h2>
            <h3 className="py-6 border-y text-2xl text-center text-gray-700">
              {configuration.MainTitle}
            </h3>

            <div>
              <Tab.Group>
                <Tab.List className="flex space-x-1 rounded-xl p-1">
                  <Tab
                    key="grouped"
                    className={({ selected }) =>
                      selected
                        ? 'w-full rounded-lg py-2.5 text-sm font-medium leading-5 text-white bg-indigo-500 shadow'
                        : 'w-full rounded-lg py-2.5 text-sm font-medium leading-5 text-gray-700 hover:bg-indigo-100 hover:text-indigo-500'
                    }>
                    Tab 1
                  </Tab>
                  <Tab
                    key="individual"
                    className={({ selected }) =>
                      selected
                        ? 'w-full rounded-lg py-2.5 text-sm font-medium leading-5 text-white bg-indigo-500 shadow'
                        : 'w-full rounded-lg py-2.5 text-sm font-medium leading-5 text-indigo-500 text-gray-600 hover:bg-indigo-100 hover:text-indigo-500'
                    }>
                    Tab 2
                  </Tab>
                </Tab.List>
                <Tab.Panel>
                  <GroupedResult />
                </Tab.Panel>
              </Tab.Group>
            </div>
          </div>
        </div>
      ) : (
        <Loading />
      )}
    </div>
  );
};

FormResult.propTypes = {
  location: PropTypes.any,
};

export default FormResult;
