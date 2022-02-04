import React from 'react';
import useChangeStatus from '../utils/useChangeStatus';
import './Status.css';
import PropTypes from 'prop-types';

/*StatusSuccess is a class that acts as a container for the display of the
status of an election */
const Status = ({status}) => {
    const {getStatus} = useChangeStatus(status);   
    return (
        <div className='status-container'>     
            {getStatus()}
        </div>
    )
}

Status.propTypes = {
    status : PropTypes.number,
}

export default Status;