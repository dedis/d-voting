import "Home.css";
import { FC } from "react";
import "i18n";
import { useTranslation } from "react-i18next";

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
