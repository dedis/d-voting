import {React, useState, useContext, useEffect} from 'react';
import './UploadFile.css';
import {Translations} from '../language/Translations';
import {LanguageContext} from '../language/LanguageContext';
import {CREATE_ENDPOINT} from '../utils/Endpoints';
import usePostCall from '../utils/usePostCall';
import PropTypes from 'prop-types';

const UploadFile = ({setShowModal, setTextModal})=> {
    const [context, ] = useContext(LanguageContext);
    const [file, setFile] = useState(null);
    const [fileExt, setFileExt] = useState(null);
    const [errors, setErrors] = useState({});
    const [name, setName] = useState('');
    const [, setIsSubmitting] = useState(false);
    const [postError, setPostError] = useState(null);
    const {postData} = usePostCall(setPostError);
    
    useEffect(()=>{
      if(postError ===null){
          setTextModal(Translations[context].electionSuccess);
      } else {
          if(postError.includes('ECONNREFUSED')){
              setTextModal(Translations[context].errorServerDown);
          } else {
              setTextModal(Translations[context].electionFail);}      
      }    
  }, [postError])

    const validateJSONFields = () => {   
      
        var data = JSON.parse(file);
        var candidates = JSON.parse(data.Format).Candidates;
        if(data.Title == ""){
          return false;
        }
        if(!Array.isArray(candidates)){
          return false;
        } else {
          /*check if the elements of the array are string*/
          for(var i = 0; i < candidates.length; i++){
            if(typeof candidates[i] !== "string"){
              return false;
            }
          }
        }
        return true;
    }

    const sendElection = async(data) => {
      let postRequest = {
        method: 'POST',
        body: JSON.stringify(data)
    }
    setPostError(null);
    postData(CREATE_ENDPOINT, postRequest, setIsSubmitting);
    }


    /*Check that the filename has indeed the extension .json
    Important: User can bypass this test by renaming the extension
     -> backend needs to perform other verification! */
    const validateFileExtension = () =>{
      let errors = {};
      if(fileExt === null){
        errors['nothing'] = Translations[context].noFile;
        setErrors(errors);
        return false;
      } else {
        let fileName = fileExt.name;
        if(fileName.substring(fileName.length-5,fileName.length)!=='.json'){
          errors['extension'] = Translations[context].notJson;
          setErrors(errors);
          return false;
        }
        return validateJSONFields();
      }    
    }

    const uploadJSON = async() => {
        if(validateFileExtension()){
          sendElection(JSON.parse(file));
          setName('');
          setShowModal(true);
        }          
    }

    const handleChange = (event) => {
      setFileExt(event.target.files[0]);
      var newUpload = event.target.files[0];
      setName(event.target.value);
      var reader = new FileReader();
      reader.onload = function(event) {
        setFile(event.target.result);
      };
     reader.readAsText(newUpload);   
    }

  return(
    <div className="form-content-right">
      <div className='option'>Option 2</div>
      {Translations[context].upload}

      <input type="file" className ='input-file'
        value = {name}
        multiple={false}
        accept='.json'
        onChange = {handleChange}  
        />
        <span className='error'>{errors.nothing}</span>
        <span className='error'>{errors.extension}</span>
        <input type="button" className = 'upload-json-btn' value={Translations[context].createElec} onClick={uploadJSON} />
    </div>
  );
}

UploadFile.propTypes = {
  setShowModal : PropTypes.func.isRequired,
  setTextModal : PropTypes.func.isRequired,
}

export default UploadFile;