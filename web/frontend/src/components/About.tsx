import { FC } from "react";
import { useTranslation } from "react-i18next";

import aboutImg from "../assets/dvoting_dela.png";

const About: FC = () => {
  const { t } = useTranslation();

  return (
    <div className="about-container">
      <div className="about-text">
        <>
          <img src={aboutImg} alt="" />
          <br />
          {t("about1")}
          <br />
          <br />
          {t("about2")}
          <br />
          <br />
          {t("about3")}
          <br />
          <br />
          {t("about4")}
          <br />
        </>
      </div>
    </div>
  );
};

export default About;
