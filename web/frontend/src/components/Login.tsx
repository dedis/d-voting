import "./Login.css";
import { FC, useState } from "react";
import "i18n";
import { useTranslation } from "react-i18next";

import { GET_TEQ_EENDPOINT } from "./utils/ExpressEndoints";
import PropTypes from "prop-types";

const Login: FC = () => {
  const { t } = useTranslation();

  const [loginError, setLoginError] = useState();

  const handleClick = async () => {
    try {
      fetch(GET_TEQ_EENDPOINT)
        .then((resp) => {
          const json_data = resp.json();
          json_data.then((result) => {
            window.location = result["url"];
          });
        })
        .catch((error) => {
          console.log(error);
        });
    } catch (error) {
      console.log(error);
    }

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
