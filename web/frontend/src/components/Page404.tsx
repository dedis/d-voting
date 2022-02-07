import React, { FC } from "react";
import { useTranslation } from "react-i18next";

const Page404: FC = () => {
  const { t } = useTranslation();

  return <div className="page404-container">{t("page404")}</div>;
};

export default Page404;
