import React, {useContext, useEffect, useState} from 'react';
import {Link} from 'react-router-dom';
import './ElectionDetails.css';
import Status from './Status';
import {Translations} from '../language/Translations';
import {LanguageContext} from '../language/LanguageContext';
import useElection from '../utils/useElection';
import Result from '../result-page/Result';
import {RESULT_AVAILABLE} from '../utils/StatusNumber';
import PropTypes from 'prop-types';
import Action from './Action';
import useGetResults from '../utils/useGetResults';

const ElectionDetails = (props) => { //props.location.data = id of the election
    const token = sessionStorage.getItem('token');
    const [context, ] = useContext(LanguageContext);   
    const {loading,title,candidates,electionID,status,result, setResult, setStatus, isResultSet, setIsResultSet} = useElection(props.location.data,token);
    const [, setError] = useState(null);  
    const [isResultAvailable, setIsResultAvailable] = useState(false); 
    const {getResults} = useGetResults();
    //fetch result when available after a status change
    useEffect(async() => {
        if(status===RESULT_AVAILABLE && isResultAvailable){
            getResults(electionID, token, setError, setResult, setIsResultSet)        
        }
    }, [status, isResultAvailable])

    return (
        <div className = 'election-details-box'>
        {!loading?
            (<div>
                <h1>{title}</h1>
                <div className='election-details-wrapper'>
                {isResultSet? <div className='election-wrapper-child'><Result resultData={result} candidates={candidates}/></div>
                    :(<div className='election-wrapper-child'> {Translations[context].status}:<Status status={status} /> 
                    <span className = 'election-action'>Action :<Action status = {status} electionID={electionID} candidates={candidates} setStatus={setStatus} setResultAvailable={setIsResultAvailable}/> </span>
                        <div className='election-candidates'>
                            {Translations[context].candidates}
                            {candidates.map((cand) => <li key={cand} className='election-candidate'>{cand}</li>)}
                        </div>      
                    </div>)}                           
                        <Link to='/elections'>
                            <button className='back-btn'>{Translations[context].back}</button>
                        </Link>               
                </div>   
            </div>)
            :(<p className='loading'>{Translations[context].loading}</p>)
        }
    </div>
    );
}

ElectionDetails.propTypes = {
    location : PropTypes.any,
}
export default ElectionDetails;