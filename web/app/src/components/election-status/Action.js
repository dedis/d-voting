
import React, {useState, useContext} from 'react';
import useChangeAction from '../utils/useChangeAction';
import './Status.css';
import Modal from '../modal/Modal';
import {Translations} from '../language/Translations';
import {LanguageContext} from '../language/LanguageContext';
import PropTypes from 'prop-types';

/**/
const Action = ({status, electionID, setStatus, setResultAvailable}) => {
    const [textModalError, setTextModalError] = useState(null);
    const [context, ] = useContext(LanguageContext);
    const [showModalError, setShowModalError] = useState(false);
    const {getAction, modalClose, modalCancel} = useChangeAction(status, electionID, setStatus, setResultAvailable, setTextModalError, setShowModalError);
    
    return (
        <span >
            {getAction()}
            {modalClose}
            {modalCancel}                   
            {<Modal showModal={showModalError} setShowModal={setShowModalError} textModal = {textModalError} buttonRightText={Translations[context].close} />}
        </span>
    )
}

Action.propTypes = {
    status : PropTypes.number,
    electionID : PropTypes.string,
    setStatus : PropTypes.func,
    setResultAvailable : PropTypes.func,
}

export default Action;