
import {React} from 'react';
import './Modal.css';
import PropTypes from 'prop-types';

const Modal = ({showModal, setShowModal, textModal,  buttonRightText}) => {
    
    const closeModal = () => {
        setShowModal(false);
    }

    const displayButtons = () => {
        return (
            <div >            
                <button className='btn-right' onClick={closeModal}>{buttonRightText}</button>              
            </div>
        )
    }

    return (
        <div>
        {showModal? (
            <div className='modal-background'>
                <div className='modal-wrapper'>
                    <div className='text-container'>{textModal}</div>          
                    <div className='buttons-container'>
                    {displayButtons()}
                    </div>
                </div>
            </div>)   
        :null}
        </div>
    );
}
Modal.propTypes = {
    showModal : PropTypes.bool.isRequired,
    setShowModal : PropTypes.func.isRequired,
    textModal: PropTypes.string.isRequired,
    buttonRightText: PropTypes.string.isRequired,
}
export default Modal;