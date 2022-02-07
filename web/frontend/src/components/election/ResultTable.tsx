import React, { FC } from "react";
import { useTranslation } from "react-i18next";

import SimpleTable from "../utils/SimpleTable";
import { RESULT_AVAILABLE } from "../utils/StatusNumber";

const ResultTable: FC = () => {
  const { t } = useTranslation();
  return (
    <div>
      <SimpleTable
        statusToKeep={RESULT_AVAILABLE}
        pathLink="results"
        textWhenData={t("displayResults")}
        textWhenNoData={t("noResultsAvailable")}
      />
    </div>
  );
};

export default ResultTable;
