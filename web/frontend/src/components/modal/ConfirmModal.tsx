import {React, useContext} from 'react';
import {Translations} from '../language/Translations';
import {LanguageContext} from '../language/LanguageContext';
import PropTypes from 'prop-types';

const ConfirmModal = ({showModal, setShowModal, textModal, setUserConfirmedAction}) => {
    const [context, ] = useContext(LanguageContext);
    
    const closeModal = () => {
        setShowModal(prev=>!prev);
    }

    const validateChoice = () => {
        setUserConfirmedAction(true);
        closeModal();
    }

    const displayButtons = () => {
        return (
            <div >
                <button className='btn-left' onClick={closeModal}>{Translations[context].no}</button>
                <button id='confirm-button' className='btn-right' onClick={validateChoice}>{Translations[context].yes}</button>             
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
ConfirmModal.propTypes = {
    showModal : PropTypes.bool.isRequired,
    setShowModal : PropTypes.func.isRequired,
    textModal: PropTypes.string.isRequired,
    setUserConfirmedAction: PropTypes.func.isRequired,
}

export default ConfirmModal;