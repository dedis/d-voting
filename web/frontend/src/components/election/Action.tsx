import { FC, useState } from "react";
import PropTypes from "prop-types";
import { useTranslation } from "react-i18next";

import useChangeAction from "../utils/useChangeAction";
import Modal from "../modal/Modal";
import "../../styles/Status.css";

type ActionProps = {
  status: number;
  electionID: string;
  setStatus(): string;
  setResultAvailable(): void;
  setTextModalError?(): string;
  setShowModal?(): void;
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

Action.propTypes = {
  status: PropTypes.number,
  electionID: PropTypes.string,
  setStatus: PropTypes.func,
  setResultAvailable: PropTypes.func,
};

export default Action;
