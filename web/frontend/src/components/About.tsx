import { FC } from "react";
import "i18n";
import { useTranslation } from "react-i18next";

import aboutImg from "../assets/dvoting_dela.png";

const About: FC = () => {
  const { t } = useTranslation();

  return (
    <div className="about-container">
      <div className="about-text">
        <img src={aboutImg} alt="" />
        {t("about")}
      </div>
    </div>
  );
};

export default About;
