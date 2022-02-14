import React, { FC, useState } from "react";
import { useTranslation } from "react-i18next";
import PropTypes from "prop-types";

import { GET_TEQ_EENDPOINT } from "../components/utils/ExpressEndoints";
import "../styles/Login.css";

const Login: FC = () => {
  const { t } = useTranslation();

  const [loginError] = useState();

  const handleClick = async () => {
    fetch(GET_TEQ_EENDPOINT)
      .then((resp) => {
        const json_data = resp.json();
        json_data.then((result) => {
          window.location = result.url;
        });
      })
      .catch((error) => {
        console.log(error);
      });

    return (
      <div>{loginError === null ? <div></div> : t("errorServerDown")}</div>
    );
  };

  return (
    <div className="login-wrapper">
      <div className="login-txt">{t("loginText")}</div>
      <button id="login-button" className="login-btn" onClick={handleClick}>
        {t("login")}
      </button>
    </div>
  );
};

Login.propTypes = {
  setToken: PropTypes.func,
};

export default Login;
