import "./BallotsTable.css";
import { FC } from "react";
import "i18n";
import { useTranslation } from "react-i18next";

import SimpleTable from "../utils/SimpleTable";
import { OPEN } from "../utils/StatusNumber";

const BallotsTable: FC = () => {
  const { t } = useTranslation();

  return (
    <div>
      <SimpleTable
        statusToKeep={OPEN}
        pathLink="vote"
        textWhenData={t("voteAllowed")}
        textWhenNoData={t("noVote")}
      />
    </div>
  );
};

export default BallotsTable;
