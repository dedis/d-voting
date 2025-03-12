import React, { FC, useContext, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { fetchCall } from 'components/utils/fetchCall';

import FormTable from './components/FormTable';
import * as endpoints from 'components/utils/Endpoints';
import Loading from 'pages/Loading';
import { LightFormInfo, Status } from 'types/form';
import FormTableFilter from './components/FormTableFilter';
import { FlashContext, FlashLevel, ProxyContext } from 'index';

const FormIndex: FC = () => {
  const { t } = useTranslation();
  const fctx = useContext(FlashContext);
  const pctx = useContext(ProxyContext);

  const [statusToKeep, setStatusToKeep] = useState<Status>(Status.Open);
  const [forms, setForms] = useState<LightFormInfo[]>(null);
  const [loading, setLoading] = useState(true);
  const [data, setData] = useState({ Forms: null });
  const [pageIndex, setPageIndex] = useState(0);
  const [error, setError] = useState(null);

  useEffect(() => {
    fetchCall(
      endpoints.forms(pctx.getProxy()),
      {
        method: 'GET',
        headers: {
          'Access-Control-Allow-Origin': '*',
        },
      },
      setData,
      setLoading
    ).catch((err) => {
      setError(err);
      setLoading(false);
    });
  }, [pctx]);

  useEffect(() => {
    if (error !== null) {
      fctx.addMessage(t('errorRetrievingForms') + error.message, FlashLevel.Error);
      setError(null);
    }
  }, [error, fctx, t]);

  // Apply the filter statusToKeep
  useEffect(() => {
    if (data.Forms === null) return;

    if (statusToKeep === null) {
      setForms(data.Forms);
      return;
    }

    const filteredForms = (data.Forms as LightFormInfo[]).filter(
      (form) => form.Status === statusToKeep
    );

    setPageIndex(0);
    setForms(filteredForms);
  }, [data, statusToKeep]);

  return (
    <div className="w-[60rem] font-sans px-4 py-4">
      {!loading ? (
        <div className="py-8">
          <h2 className="pb-2 text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
            {t('forms')}
          </h2>
          <div className="mt-1 flex flex-col sm:mt-0">
            <div className="mt-2 flex items-center text-sm text-gray-500">{t('listForm')}</div>
            <div className="mt-1 flex items-center text-sm text-gray-500">{t('clickForm')}</div>
          </div>

          <FormTableFilter setStatusToKeep={setStatusToKeep} />
          <FormTable pageIndex={pageIndex} setPageIndex={setPageIndex} forms={forms} />
        </div>
      ) : (
        <Loading />
      )}
    </div>
  );
};

export default FormIndex;
