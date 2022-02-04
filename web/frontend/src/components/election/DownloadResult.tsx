import React, {useContext} from 'react';
import {saveAs} from 'file-saver';
import {Translations} from '../language/Translations';
import {LanguageContext} from '../language/LanguageContext';
import PropTypes from 'prop-types';

const DownloadResult = ({resultData}) => {
    const [context, ] = useContext(LanguageContext);
    const fileName = 'result.json';

    // Create a blob of the data
    const fileToSave = new Blob([JSON.stringify({Result: resultData})], {
        type: 'application/json',
        name: fileName
    });

    const handleClick = () => {
        saveAs(fileToSave, fileName);
    }

    return( <div>
                <button className='back-btn' onClick={handleClick}>{Translations[context].download}</button>
            </div>
    );
}
DownloadResult.propTypes = {
    resultData : PropTypes.object,
}
export default DownloadResult;

//https://stackoverflow.com/questions/19721439/download-json-object-as-a-file-from-browser