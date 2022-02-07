import React, { useState, useEffect } from "react";
import PropTypes from "prop-types";
import { useTranslation } from "react-i18next";

import { CREATE_ENDPOINT } from "../utils/Endpoints";
import usePostCall from "../utils/usePostCall";

const UploadFile = ({ setShowModal, setTextModal }) => {
  const { t } = useTranslation();
  const [file, setFile] = useState(null);
  const [fileExt, setFileExt] = useState(null);
  const [errors, setErrors] = useState({ nothing: "", extension: "" });
  const [name, setName] = useState("");
  const [, setIsSubmitting] = useState(false);
  const [postError, setPostError] = useState(null);
  const { postData } = usePostCall(setPostError);

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
  }, [postError]);

  const validateJSONFields = () => {
    var data = JSON.parse(file);
    var candidates = JSON.parse(data.Format).Candidates;
    if (data.Title == "") {
      return false;
    }
    if (!Array.isArray(candidates)) {
      return false;
    } else {
      /*check if the elements of the array are string*/
      for (var i = 0; i < candidates.length; i++) {
        if (typeof candidates[i] !== "string") {
          return false;
        }
      }
    }
    return true;
  };

  const sendElection = async (data) => {
    let postRequest = {
      method: "POST",
      body: JSON.stringify(data),
    };
    setPostError(null);
    postData(CREATE_ENDPOINT, postRequest, setIsSubmitting);
  };

  /*Check that the filename has indeed the extension .json
  Important: User can bypass this test by renaming the extension
   -> backend needs to perform other verification! */
  const validateFileExtension = () => {
    if (fileExt === null) {
      errors.nothing = t("noFile");
      setErrors(errors);
      return false;
    } else {
      let fileName = fileExt.name;
      if (
        fileName.substring(fileName.length - 5, fileName.length) !== ".json"
      ) {
        errors.extension = t("notJson");
        setErrors(errors);
        return false;
      }
      return validateJSONFields();
    }
  };

  const uploadJSON = async () => {
    if (validateFileExtension()) {
      sendElection(JSON.parse(file));
      setName("");
      setShowModal(true);
    }
  };

  const handleChange = (event) => {
    setFileExt(event.target.files[0]);
    var newUpload = event.target.files[0];
    setName(event.target.value);
    var reader = new FileReader();
    reader.onload = function (event) {
      setFile(event.target.result);
    };
    reader.readAsText(newUpload);
  };

  return (
    <div className="form-content-right bg-gray-200 flex-1 m-1 p-10">
      <div className="uppercase font-bold py-5">Option 2</div>
      {t("upload")}

      <input
        type="file"
        className="block bg-transparent hover:bg-blue-500 text-blue-700 font-semibold hover:text-white text-xs py-1 px-4 border border-blue-500 hover:border-transparent rounded"
        value={name}
        multiple={false}
        accept=".json"
        onChange={handleChange}
      />
      <span className="error">{errors.nothing}</span>
      <span className="error">{errors.extension}</span>
      <input
        type="button"
        className="block bg-blue-500 hover:bg-blue-700 text-white font-bold mt-7 py-2 px-5 rounded-full text-xs"
        value={t("createElec") as string}
        onClick={uploadJSON}
      />
    </div>
  );
};

UploadFile.propTypes = {
  setShowModal: PropTypes.func.isRequired,
  setTextModal: PropTypes.func.isRequired,
};

export default UploadFile;
