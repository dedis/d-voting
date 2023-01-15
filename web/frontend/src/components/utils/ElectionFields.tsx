import { LightFormInfo } from 'types/form';
import { useFillLightFormInfo } from './FillFormInfo';

/**
 *
 * @param {*} formData a json object of a form
 * @returns the fields of a form and a function to change the status field
 */
const FormFields = (formData: LightFormInfo) => {
  const { title, titleFr, titleDe,id, status, pubKey, setStatus } = useFillLightFormInfo(formData);
  return { title, id, status, pubKey, setStatus };
};

export default FormFields;
