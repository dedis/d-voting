import React from 'react';
import PropTypes from 'prop-types';
import useChangeStatus from './utils/useChangeStatus';

// FormStatus is a class that acts as a container for the display of the
// status of an form
const FormStatus = ({ status }) => {
  const { getStatus } = useChangeStatus(status);
  return <div className="inline-block align-left">{getStatus()}</div>;
};

FormStatus.propTypes = {
  status: PropTypes.number,
};

export default FormStatus;
