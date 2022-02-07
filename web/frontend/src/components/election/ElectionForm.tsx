import React, { FC, useState, useEffect } from "react";
import PropTypes from "prop-types";
import { useTranslation } from "react-i18next";

import { CREATE_ENDPOINT } from "../utils/Endpoints";
import usePostCall from "../utils/usePostCall";
import {
  COLLECTIVE_AUTHORITY_MEMBERS,
  SHUFFLE_THRESHOLD,
} from "../utils/CollectiveAuthorityMembers";

type ElectionFormProps = {
  setShowModal(modal: any): void;
  setTextModal(text: string): void;
};

const ElectionForm: FC<ElectionFormProps> = ({
  setShowModal,
  setTextModal,
}) => {
  const { t } = useTranslation();

  const [electionName, setElectionName] = useState("");
  const [newCandidate, setNewCandidate] = useState("");
  const [candidates, setCandidates] = useState([]);
  const [errors, setErrors] = useState({
    unique: "",
    empty: "",
    newCandidate: "",
    candidates: "",
  });
  const [postError, setPostError] = useState(null);
  const { postData } = usePostCall(setPostError);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    if (postError === null) {
      setTextModal(t("electionSuccess"));
    } else {
      if (postError.includes("ECONNREFUSED")) {
        setTextModal(t("errorServerDown"));
      } else {
        setTextModal(t("electionFail"));
      }
    }
  }, [isSubmitting]);

  const sendFormData = async () => {
    //create the JSON object
    const election = {};
    election["Title"] = electionName;
    election["AdminID"] = sessionStorage.getItem("id");
    election["ShuffleThreshold"] = SHUFFLE_THRESHOLD;
    election["Members"] = COLLECTIVE_AUTHORITY_MEMBERS;
    election["Format"] = JSON.stringify({ Candidates: candidates });
    election["Token"] = sessionStorage.getItem("token");
    console.log(JSON.stringify(election));
    let postRequest = {
      method: "POST",
      body: JSON.stringify(election),
    };
    setPostError(null);
    postData(CREATE_ENDPOINT, postRequest, setIsSubmitting);
  };

  const validate = () => {
    let isValid = true;

    if (candidates.length === 0) {
      errors.candidates = t("errorCandidates");
      isValid = false;
    }
    if (newCandidate.length !== 0) {
      errors.newCandidate = t("errorNewCandidate") + newCandidate + "?";
      isValid = false;
    }
    setErrors(errors);
    return isValid;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (validate()) {
      await sendFormData();
      setShowModal((prev) => !prev);
      setElectionName("");
      setNewCandidate("");
      setCandidates([]);
      setPostError(null);
    }
  };

  const onSubmitPreventDefault = (e) => {
    e.preventDefault();
  };

  const handleChangeName = (e) => {
    setElectionName(e.target.value);
  };

  const handleChangeCandidate = (e) => {
    e.preventDefault();
    setNewCandidate(e.target.value);
  };

  const isCandidateUnique = (cand) => {
    return !candidates.some((item) => cand === item);
  };

  const handleAdd = (e) => {
    e.preventDefault();
    if (newCandidate.length === 0) {
      errors.empty = t("nothingToAdd");
      setErrors(errors);
      return;
    }

    if (!isCandidateUnique(newCandidate)) {
      errors.unique = t("duplicateCandidate");
      setErrors(errors);
      setNewCandidate("");
      return;
    }

    setNewCandidate("");
    errors.newCandidate = "";
    setErrors(errors);
    setCandidates(candidates.concat(newCandidate));
  };

  const handleDelete = (cand) => {
    const choices = candidates.filter((candi) => candi !== cand);
    setCandidates(choices);
  };

  const handleKeyPress = (e) => {
    if (e.key === "Enter") {
      e.preventDefault();
      handleAdd(e);
    }
  };

  const handleKeyPressTitle = (e) => {
    if (e.key === "Enter") {
      e.preventDefault();
    }
  };

  return (
    <div className="form-wrapper bg-gray-200 flex-1 m-1 p-10">
      <div className="uppercase font-bold py-5">Option 1</div>

      <form className="form-choices" onSubmit={handleSubmit}>
        <div className="flex flex-wrap -mx-3 mb-6">
          <div className="w-full px-3">
            <label
              className="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
              htmlFor="new-name"
            >
              {t("elecName")}*:
            </label>
            <input
              required
              id="new-name"
              type="text"
              placeholder={t("namePlaceHolder")}
              value={electionName}
              onChange={handleChangeName}
              onKeyPress={handleKeyPressTitle}
              className="appearance-none block w-full text-gray-700 border border-gray-200 rounded py-3 px-4 mb-3 leading-tight focus:outline-none bg-white focus:border-gray-500"
            />
          </div>
        </div>

        <div className="flex flex-wrap -mx-3 mb-6">
          <div className="w-full px-3">
            <label
              className="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
              htmlFor="new-choice"
            >
              {t("addCandidate")} *:
            </label>
            <input
              id="new-choice"
              type="text"
              placeholder={t("addCandPlaceHolder")}
              name="newCandidate"
              value={newCandidate}
              onChange={handleChangeCandidate}
              onKeyPress={handleKeyPress}
              onSubmit={handleAdd}
              className="appearance-none block w-full text-gray-700 border border-gray-200 rounded py-3 px-4 mb-3 leading-tight focus:outline-none bg-white focus:border-gray-500"
            />
          </div>

          <button
            type="button"
            className="bg-transparent hover:bg-blue-500 text-blue-700 font-semibold hover:text-white text-xs ml-4 py-1 px-4 border border-blue-500 hover:border-transparent rounded"
            onClick={handleAdd}
            onSubmit={onSubmitPreventDefault}
          >
            {t("add")}
          </button>
          <div className="form-error">{errors.unique}</div>
          <div className="form-error">{errors.empty}</div>
          <div className="form-error">{errors.newCandidate}</div>
          <div className="form-error">{errors.candidates}</div>
        </div>

        <div className="form-candidates">
          <ul className="choices-saved">
            {candidates.map((cand, i) => (
              <div key={i}>
                <li className="text-sm">
                  {cand}
                  <button
                    type="button"
                    className="font-semibold pl-4"
                    onClick={() => handleDelete(cand)}
                    onSubmit={onSubmitPreventDefault}
                  >
                    {t("delete")}
                  </button>
                </li>
              </div>
            ))}
          </ul>
        </div>
        <div>
          <button
            type="submit"
            className="bg-blue-500 hover:bg-blue-700 text-white font-bold mt-7 py-2 px-5 rounded-full text-xs"
            onSubmit={handleSubmit}
          >
            {t("createElec")}
          </button>
        </div>
      </form>
    </div>
  );
};
ElectionForm.propTypes = {
  setShowModal: PropTypes.func.isRequired,
  setTextModal: PropTypes.func.isRequired,
};

export default ElectionForm;
