import React, { FC } from 'react';
import { saveAs } from 'file-saver';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

type DownloadResultProps = {
  resultData: string;
};

const DownloadResult: FC<DownloadResultProps> = (resultData) => {
  const { t } = useTranslation();
  const fileName = 'result.json';

  // Create a blob of the data
  const fileToSave = new Blob([JSON.stringify({ Result: resultData })], {
    type: 'application/json',
  });

  const handleClick = () => {
    saveAs(fileToSave, fileName);
  };

  return (
    <div>
      <button className="back-btn" onClick={handleClick}>
        {t('download')}
      </button>
    </div>
  );
};

DownloadResult.propTypes = {
  resultData: PropTypes.string.isRequired,
};

export default DownloadResult;

//https://stackoverflow.com/questions/19721439/download-json-object-as-a-file-from-browser
