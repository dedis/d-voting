import{React, useContext} from 'react';
import useElection from '../utils/useElection';
import Result from './Result';
import {Translations} from '../language/Translations';
import {LanguageContext} from '../language/LanguageContext';
import {Link} from 'react-router-dom';
import PropTypes from 'prop-types';
import './ResultPage.css';

const ResultPage = (props) =>{ //props.location.data = id of the election
    const token = sessionStorage.getItem('token');
    const [context, ] = useContext(LanguageContext);   
    const {loading,title,candidates,result, error} = useElection(props.location.data,token);

    return(<div className='result-box'>
    {!loading?
        (<div>
            <h1>{title}</h1>
            <Result resultData={result} candidates={candidates}/>
        </div>)   
        : 
        (error===null?<p className='loading'>{Translations[context].loading} </p>:<div className='error-retrieving'>{Translations[context].errorRetrievingElection}</div>)}
       <Link to='/results'>
                            <button className='back-btn'>{Translations[context].back}</button>
                        </Link>  
        
    </div>)

}
ResultPage.propTypes = {
    location : PropTypes.any,
}
export default ResultPage;