import React, { FC, useEffect, useState } from 'react';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';
import { LightFormInfo } from 'types/form';
import FormRow from './FormRow';

type FormTableProps = {
  forms: LightFormInfo[];
  setPageIndex: (index: number) => void;
  pageIndex: number;
};

// Returns a table where each line corresponds to an form with
// its name, status and a quickAction if available
const FORM_PER_PAGE = 10;

const FormTable: FC<FormTableProps> = ({ forms, pageIndex, setPageIndex }) => {
  const { t } = useTranslation();
  const [formsToDisplay, setFormsToDisplay] = useState<LightFormInfo[]>([]);

  const partitionArray = (array: LightFormInfo[], size: number) => {
    if (array !== null) {
      return array
        .map((_v, i) => (i % size === 0 ? array.slice(i, i + size) : null))
        .filter((v) => v);
    }

    return [];
  };

  useEffect(() => {
    if (forms !== null) {
      setFormsToDisplay(partitionArray(forms, FORM_PER_PAGE)[pageIndex]);
    }
  }, [pageIndex, forms]);

  const handlePrevious = (): void => {
    if (pageIndex > 0) {
      setPageIndex(pageIndex - 1);
    }
  };

  const handleNext = (): void => {
    if (partitionArray(forms, FORM_PER_PAGE).length > pageIndex + 1) {
      setPageIndex(pageIndex + 1);
    }
  };

  return (
    <div>
      <div className="relative divide-y overflow-x-auto shadow-md sm:rounded-lg">
        <table className="w-full text-sm text-left text-gray-500">
          <thead className="text-xs text-gray-700 uppercase bg-gray-50">
            <tr>
              <th scope="col" className="px-1.5 sm:px-6 py-3">
                {t('elecName')}
              </th>
              <th scope="col" className="px-1.5 sm:px-6 py-3">
                {t('status')}
              </th>
              <th scope="col" className="px-1.5 sm:px-6 py-3">
                <span className="sr-only">Edit</span>
              </th>
            </tr>
          </thead>
          <tbody>
            {formsToDisplay !== undefined &&
              formsToDisplay.map((form) => (
                <React.Fragment key={form.FormID}>
                  <FormRow form={form} />
                </React.Fragment>
              ))}
          </tbody>
        </table>
        <nav
          className="bg-white px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6"
          aria-label="Pagination">
          <div className="hidden sm:block text-sm text-gray-700">
            {t('showingNOverMOfXResults', {
              n: pageIndex + 1,
              m: partitionArray(forms, FORM_PER_PAGE).length,
              x: `${forms !== null ? forms.length : 0}`,
            })}
          </div>
          <div className="flex-1 flex justify-between sm:justify-end">
            <button
              disabled={pageIndex === 0}
              onClick={handlePrevious}
              className="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
              {t('previous')}
            </button>
            <button
              disabled={partitionArray(forms, FORM_PER_PAGE).length <= pageIndex + 1}
              onClick={handleNext}
              className="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
              {t('next')}
            </button>
          </div>
        </nav>
      </div>
    </div>
  );
};

FormTable.propTypes = {
  forms: PropTypes.array,
};

export default FormTable;
