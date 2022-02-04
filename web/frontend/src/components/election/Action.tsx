import React, { FC, useState } from "react";
import PropTypes from "prop-types";
import { useTranslation } from "react-i18next";

import useChangeAction from "../utils/useChangeAction";
import Modal from "../modal/Modal";
import "../../styles/Status.css";

type ActionProps = {
  status: typeof PropTypes.number;
  electionID: typeof PropTypes.string;
  setStatus: typeof PropTypes.func;
  setResultAvailable: typeof PropTypes.func;
  setTextModalError?: typeof PropTypes.func;
  setShowModal?: typeof PropTypes.func;
};

/**/
const Action: FC<ActionProps> = ({
  status,
  electionID,
  setStatus,
  setResultAvailable,
}) => {
  const { t } = useTranslation();

  const [textModalError, setTextModalError] = useState(null);
  const [showModalError, setShowModalError] = useState(false);
  const { getAction, modalClose, modalCancel } = useChangeAction(
    status,
    electionID,
    setStatus,
    setResultAvailable,
    setTextModalError,
    setShowModalError
  );

  return (
    <span>
      {getAction()}
      {modalClose}
      {modalCancel}
      {
        <Modal
          showModal={showModalError}
          setShowModal={setShowModalError}
          textModal={textModalError === null ? "" : textModalError}
          buttonRightText={t("close")}
        />
      }
    </span>
  );
};

export default Action;
