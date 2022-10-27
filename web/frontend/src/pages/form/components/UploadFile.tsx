import configurationSchema from '../../../schema/configurationValidation';

import Ajv from 'ajv';

import configurationJSONSchema from '../../../schema/form_conf.json';
import { unmarshalConfig } from '../../../types/JSONparser';
import { Configuration } from 'types/configuration';
import { useTranslation } from 'react-i18next';

const ajv = new Ajv({
  schemas: [configurationJSONSchema],
});

const UploadFile = ({ updateForm, setShowModal, setTextModal }) => {
  const { t } = useTranslation();
  const validateJSONSchema = ajv.getSchema('configurationSchema');

  const handleDrop = (file) => {
    if (!file || file.type !== 'application/json') {
      return;
    }

    var reader = new FileReader();

    reader.onload = async function (param) {
      const result: string = JSON.parse(param.target.result.toString());
      if (!validateJSONSchema(result)) {
        setTextModal('Invalid schema JSON file');
        setShowModal(true);
        return;
      }

      try {
        const validConf: any = await configurationSchema.validate(result);
        // unmarshal the configuration to add the Types on the objects
        const unmarshalConfigResult: Configuration = unmarshalConfig(validConf);
        updateForm(unmarshalConfigResult);
      } catch (err) {
        setTextModal('Incorrect form configuration : ' + err.errors.join(','));
        setShowModal(true);
      }
    };

    reader.readAsText(file);
  };

  return (
    <div className="px-4 py-4">
      {t('createElecDesc')}
      <input
        className="hidden"
        type={'file'}
        accept=".json"
        id="uploadJSON"
        onChange={(e) => handleDrop(e.target.files[0])}
      />
      <label
        className="ml-1 cursor-pointer font-medium text-indigo-600 hover:text-indigo-500"
        htmlFor="uploadJSON">
        {t('uploadJSON')}
      </label>
    </div>
  );
};

export default UploadFile;
