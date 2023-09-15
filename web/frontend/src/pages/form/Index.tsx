import React, { FC, useContext, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import FormTable from './components/FormTable';
import useFetchCall from 'components/utils/useFetchCall';
import * as endpoints from 'components/utils/Endpoints';
import Loading from 'pages/Loading';
import { LightFormInfo, Status } from 'types/form';
import FormTableFilter from './components/FormTableFilter';
import { FlashContext, FlashLevel, ProxyContext } from 'index';

const FormIndex: FC = () => {
  const { t } = useTranslation();
  const fctx = useContext(FlashContext);
  const pctx = useContext(ProxyContext);

  const [statusToKeep, setStatusToKeep] = useState<Status>(null);
  const [forms, setForms] = useState<LightFormInfo[]>(null);
  const [loading, setLoading] = useState(true);
  const [pageIndex, setPageIndex] = useState(0);

  const request = {
    method: 'GET',
    headers: {
      'Access-Control-Allow-Origin': '*',
    },
  };

  const [data, dataLoading, error] = useFetchCall(endpoints.forms(pctx.getProxy()), request);

  useEffect(() => {
    if (error !== null) {
      fctx.addMessage(t('errorRetrievingForms') + error.message, FlashLevel.Error);
      setLoading(false);
    }
  }, [fctx, t, error]);

  // Apply the filter statusToKeep
  useEffect(() => {
    if (data === null) return;

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

  useEffect(() => {
    if (dataLoading !== null) {
      setLoading(dataLoading);
    }
  }, [dataLoading]);

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
