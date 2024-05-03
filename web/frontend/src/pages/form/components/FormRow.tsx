import React, { FC, useContext, useEffect, useState } from 'react';
import { LightFormInfo } from 'types/form';
import { Link } from 'react-router-dom';
import FormStatus from './FormStatus';
import QuickAction from './QuickAction';
import { default as i18n } from 'i18next';
import { AuthContext } from '../../..';

type FormRowProps = {
  form: LightFormInfo;
};

const SUBJECT_ELECTION = 'election';
const ACTION_CREATE = 'create';

const FormRow: FC<FormRowProps> = ({ form }) => {
  const [titles, setTitles] = useState<any>({});
  const authCtx = useContext(AuthContext);
  useEffect(() => {
    if (form.Title === undefined) return;
    setTitles({ En: form.Title.En, Fr: form.Title.Fr, De: form.Title.De, URL: form.Title.URL });
  }, [form]);
  // let i18next handle choosing the appropriate language
  const formRowI18n = i18n.createInstance();
  formRowI18n.init();
  // get current language
  formRowI18n.changeLanguage(i18n.language);
  Object.entries(titles).forEach(([lang, title]: [string, string | undefined]) => {
    if (title) {
      formRowI18n.addResource(lang.toLowerCase(), 'form', 'title', title);
    }
  });
  const formTitle = formRowI18n.t('title', { ns: 'form', fallbackLng: 'en' });
  return (
    <tr className="bg-white border-b hover:bg-gray-50 ">
      <td className="px-1.5 sm:px-6 py-4 font-medium text-gray-900 whitespace-nowrap truncate">
        {authCtx.isLogged && authCtx.isAllowed(SUBJECT_ELECTION, ACTION_CREATE) ? (
          <Link className="text-gray-700 hover:text-[#ff0000]" to={`/forms/${form.FormID}`}>
            <div className="max-w-[20vw] truncate">{formTitle}</div>
          </Link>
        ) : (
          <div className="max-w-[20vw] truncate">{formTitle}</div>
        )}
      </td>
      <td className="px-1.5 sm:px-6 py-4">{<FormStatus status={form.Status} />}</td>
      <td className="px-1.5 sm:px-6 py-4 text-right">
        <QuickAction status={form.Status} formID={form.FormID} />
      </td>
    </tr>
  );
};

export default FormRow;
