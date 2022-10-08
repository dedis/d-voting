import { LightFormInfo } from 'types/form';
import { useFillLightFormInfo } from './FillFormInfo';

/**
 *
 * @param {*} formData a json object of an form
 * @returns the fields of an form and a function to change the status field
 */
const FormFields = (formData: LightFormInfo) => {
  const { title, id, status, pubKey, setStatus } = useFillLightFormInfo(formData);
  return { title, id, status, pubKey, setStatus };
};

export default FormFields;
