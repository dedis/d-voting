import React, { FC } from 'react';
import { LightFormInfo } from 'types/form';
import { Link } from 'react-router-dom';
import FormStatus from './FormStatus';
import QuickAction from './QuickAction';

type FormRowProps = {
  form: LightFormInfo;
};

const FormRow: FC<FormRowProps> = ({ form }) => {
  return (
    <tr className="bg-white border-b hover:bg-gray-50 ">
      <td className="px-1.5 sm:px-6 py-4 font-medium text-gray-900 whitespace-nowrap truncate">
        <Link className="text-gray-700 hover:text-indigo-500" to={`/forms/${form.FormID}`}>
          <div className="max-w-[20vw] truncate">{form.Title}</div>
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
