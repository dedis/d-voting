import { useDropzone } from 'react-dropzone';
import configurationSchema from '../../../schema/configurationValidation';

import Ajv from 'ajv';

import configurationJSONSchema from '../../../schema/election_conf.json';
import { marshallConfig } from './utils/JSONparser';
import { Configuration } from 'types/configuration';

const ajv = new Ajv({
  schemas: [configurationJSONSchema],
});

const UploadFile = ({ setConf, setShowModal, setTextModal }) => {
  const validateJSONSchema = ajv.getSchema('configurationSchema');

  const handleDrop = (file) => {
    if (file && file.type === 'application/json') {
      var reader = new FileReader();
      reader.onload = async function (param) {
        const result: any = JSON.parse(param.target.result.toString());
        if (validateJSONSchema(result)) {
          try {
            const validConf: any = await configurationSchema.validate(result);
            // marshall the configuration to add the Types on the objects
            const marshallConfigResult: Configuration = marshallConfig(validConf);
            setConf(marshallConfigResult);
          } catch (err) {
            setTextModal('Incorrect election configuration : ' + err.errors.join(','));
            setShowModal(true);
          }
        } else {
          setTextModal('Invalid schema JSON file');
          setShowModal(true);
        }
      };
      reader.readAsText(file);
    }
  };

  const { getRootProps, getInputProps } = useDropzone({
    onDrop: (files) => handleDrop(files[0]),
    accept: '.json',
    maxFiles: 1,
  });

  return (
    <div className="py-2 px-4">
      <div className="mt-1 flex justify-center px-6 pt-5 pb-6 border-2 border-gray-300 border-dashed rounded-md">
        <div {...getRootProps({ className: 'dropzone' })} className="space-y-1 text-center">
          <svg
            className="mx-auto h-12 w-12 text-gray-400"
            stroke="currentColor"
            fill="none"
            viewBox="0 0 48 48"
            aria-hidden="true">
            <path
              d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8m-12 4h.02"
              strokeWidth={2}
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
          <div className="flex text-sm text-gray-600">
            <div className="relative cursor-pointer bg-white rounded-md font-medium text-indigo-600 hover:text-indigo-500">
              <span>Upload a file</span>
              <input {...getInputProps()} />
            </div>
            <p className="pl-1">or drag and drop</p>
          </div>
          <p className="text-xs text-gray-500">JSON file only</p>
        </div>
      </div>
    </div>
  );
};

export default UploadFile;
