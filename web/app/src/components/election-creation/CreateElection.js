import {React,useContext, useState} from 'react';

import './CreateElection.css';
import ElectionForm from './ElectionForm.js'
import UploadFile from './UploadFile';
import {Translations} from '../language/Translations';
import {LanguageContext} from '../language/LanguageContext';
import Modal from '../modal/Modal';


const CreateElection = () => {
  const [context, ] = useContext(LanguageContext);
  const [showModal, setShowModal] = useState(false);
  const [textModal, setTextModal] = useState('');

  return (
    <div className= 'create-election-wrapper'>
      <Modal showModal={showModal} setShowModal={setShowModal} textModal = {textModal} buttonRightText={Translations[context].close} />     
      <h4>{Translations[context].create}</h4>
      
      <div className='election-form'>
        <ElectionForm setShowModal={setShowModal} setTextModal={setTextModal} />     
        <UploadFile setShowModal={setShowModal} setTextModal={setTextModal}/>
      </div>
    </div>
  );
}

export default CreateElection;
