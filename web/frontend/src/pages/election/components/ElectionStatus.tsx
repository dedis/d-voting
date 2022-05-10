import React from 'react';
import PropTypes from 'prop-types';

import useChangeStatus from 'components/utils/useChangeStatus';

// ElectionStatus is a class that acts as a container for the display of the
// status of an election
const ElectionStatus = ({ status }) => {
  const { getStatus } = useChangeStatus(status);
  return <div className="inline-block align-left">{getStatus()}</div>;
};

ElectionStatus.propTypes = {
  status: PropTypes.number,
};

export default ElectionStatus;
