import React, { FC, useEffect, useState } from 'react';
import { LightFormInfo } from 'types/form';
import { Link } from 'react-router-dom';
import FormStatus from './FormStatus';
import QuickAction from './QuickAction';
import { default as i18n } from 'i18next';
import { isJson } from 'types/JSONparser';

type FormRowProps = {
  form: LightFormInfo;
};

const FormRow: FC<FormRowProps> = ({ form }) => {
  const [titles, setTitles] = useState<any>({});
  useEffect(() => {
    if (form.Title === '') return;
    if (isJson(form.Title)) {
      setTitles(JSON.parse(form.Title));
    } else {
      setTitles({ en: form.Title, fr: form.TitleFr, de: form.TitleDe });
    }
  }, [form]);
  // let i18next handle choosing the appropriate language
  const formRowI18n = i18n.createInstance();
  formRowI18n.init();
  // get current language
  formRowI18n.changeLanguage(i18n.language);
  Object.entries(titles).forEach(([lang, title]: [string, string | undefined]) => {
    if (title) {
      formRowI18n.addResource(lang, 'form', 'title', title);
    }
  });
  return (
    <tr className="bg-white border-b hover:bg-gray-50 ">
      <td className="px-1.5 sm:px-6 py-4 font-medium text-gray-900 whitespace-nowrap truncate">
        <Link className="text-gray-700 hover:text-indigo-500" to={`/forms/${form.FormID}`}>
          <div className="max-w-[20vw] truncate">
            {formRowI18n.t('title', { ns: 'form', fallbackLng: 'en' })}
          </div>
        </Link>
      </td>
      <td className="px-1.5 sm:px-6 py-4">{<FormStatus status={form.Status} />}</td>
      <td className="px-1.5 sm:px-6 py-4 text-right">
        <QuickAction status={form.Status} formID={form.FormID} />
      </td>
    </tr>
  );
};

export default FormRow;
