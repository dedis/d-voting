import { FC } from 'react';
import { useTranslation } from 'react-i18next';

import FormForm from './components/FormForm';

const FormCreate: FC = () => {
  const { t } = useTranslation();

  return (
    <div className="w-[70rem] font-sans md:px-4 md:py-4">
      <div className="px-4 pt-4">
        <h2 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
          {t('navBarCreate')}
        </h2>
      </div>
      <FormForm />
    </div>
  );
};

export default FormCreate;
