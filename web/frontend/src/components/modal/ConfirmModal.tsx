import React, { FC } from "react";
import PropTypes from "prop-types";
import { useTranslation } from "react-i18next";

type ConfirmModalProps = {
  showModal: boolean;
  setShowModal: (prev: boolean) => void;
  textModal: string;
  setUserConfirmedAction: (confirmed: boolean) => void;
};

const ConfirmModal: FC<ConfirmModalProps> = ({
  showModal,
  setShowModal,
  textModal,
  setUserConfirmedAction,
}) => {
  const { t } = useTranslation();

  const closeModal = () => {
    setShowModal((prev) => !prev);
  };

  const validateChoice = () => {
    setUserConfirmedAction(true);
    closeModal();
  };

  const displayButtons = () => {
    return (
      <div>
        <button className="btn-left" onClick={closeModal}>
          {t("no")}
        </button>
        <button
          id="confirm-button"
          className="btn-right"
          onClick={validateChoice}
        >
          {t("yes")}
        </button>
      </div>
    );
  };

  return (
    <div>
      {showModal ? (
        <div className="modal-background">
          <div className="modal-wrapper">
            <div className="text-container">{textModal}</div>
            <div className="buttons-container">{displayButtons()}</div>
          </div>
        </div>
      ) : null}
    </div>
  );
};

ConfirmModal.propTypes = {
  showModal: PropTypes.bool.isRequired,
  setShowModal: PropTypes.func.isRequired,
  textModal: PropTypes.string.isRequired,
  setUserConfirmedAction: PropTypes.func.isRequired,
};

export default ConfirmModal;
