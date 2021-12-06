import { React, useContext, useState } from 'react';

import ElectionForm from './ElectionForm.js'
import UploadFile from './UploadFile';
import { Translations } from '../language/Translations';
import { LanguageContext } from '../language/LanguageContext';
import Modal from '../modal/Modal';


const CreateElection = () => {
  const [context,] = useContext(LanguageContext);
  const [showModal, setShowModal] = useState(false);
  const [textModal, setTextModal] = useState('');

  return (
    <div className='create-election-wrapper'>
      <Modal showModal={showModal} setShowModal={setShowModal} textModal={textModal} buttonRightText={Translations[context].close} />
      <h1 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
        {Translations[context].navBarCreate}
      </h1>
      <p className="py-5">{Translations[context].create}</p>

      <div className='election-form flex flex-row justify-center items-start'>
        <ElectionForm setShowModal={setShowModal} setTextModal={setTextModal} />
        <UploadFile setShowModal={setShowModal} setTextModal={setTextModal} />
      </div>
    </div>
  );
}

export default CreateElection;
