import {React,useContext} from 'react';
import {Translations} from '../language/Translations';
import {LanguageContext} from '../language/LanguageContext';
import {OPEN, CLOSED, SHUFFLED_BALLOT, RESULT_AVAILABLE, CANCELED} from '../utils/StatusNumber';

/*Custom hook that can display the status of an election and enable changes of status (closing, cancelling,...)*/ 
const useChangeStatus = (status) =>{
    const [context, ] = useContext(LanguageContext);

    const getStatus = () => {
        switch (status){     
            case OPEN: 
                return <span className = 'election-status'>
                    <span className='election-status-on'></span>
                    <span className='election-status-text'>{Translations[context].statusOpen}</span>
               </span> 
            case CLOSED: 
                return  <span className = 'election-status'>               
                            <span className='election-status-closed'></span>
                            <span className='election-status-text'>{Translations[context].statusClose}</span>                           
                        </span>; 
            case SHUFFLED_BALLOT: 
                return <span className = 'election-status'>                   
                            <span className='election-status-closed'></span>
                            <span className='election-status-text'>{Translations[context].statusShuffle}</span>                     
                        </span>;
            case RESULT_AVAILABLE: 
                return <span className = 'election-status'>
                        <span className='election-status-closed'></span>
                        <span className='election-status-text'>{Translations[context].resultsAvailable }</span>                    
                     </span>;               
            case CANCELED: 
                return <span className = 'election-status'>
                    <span className='election-status-cancelled'></span>
                    <span className='election-status-text'>{Translations[context].statusCancel}</span>
                </span>;  
            default :
                return null
        }
    } 
    return {getStatus};
};

export default useChangeStatus;