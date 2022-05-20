import PropTypes from 'prop-types';
import useChangeDKGStatus from './utils/useChangeDKGStatus';

// DKGStatus is a class that acts as a container for the display of the
// status of a node
const DKGStatus = ({ status }) => {
  const { getDKGStatus } = useChangeDKGStatus(status);
  return <div className="inline-block align-left">{getDKGStatus()}</div>;
};

DKGStatus.propTypes = {
  status: PropTypes.number,
};

export default DKGStatus;
