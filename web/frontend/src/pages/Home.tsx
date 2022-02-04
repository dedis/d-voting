import React, { FC } from "react";
import { useTranslation } from "react-i18next";

import "../styles/Home.css";

const Home: FC = () => {
  const { t } = useTranslation();

  return (
    <div className="home">
      <h1>{t("homeTitle")}</h1>
      <div className="home-txt">{t("homeText")}</div>
    </div>
  );
};

export default Home;
