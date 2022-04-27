import React from 'react';
import PropTypes from 'prop-types';

import useChangeStatus from 'components/utils/useChangeStatus';

/*StatusSuccess is a class that acts as a container for the display of the
status of an election */
const Status = ({ status }) => {
  const { getStatus } = useChangeStatus(status);
  return <div className="inline-block align-left">{getStatus()}</div>;
};

Status.propTypes = {
  status: PropTypes.number,
};

export default Status;
