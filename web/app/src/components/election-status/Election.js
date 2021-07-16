import React, {useContext} from 'react';
import './Election.css';
import ElectionTable from './ElectionTable';
import {Translations} from '../language/Translations';
import {LanguageContext} from '../language/LanguageContext';
import useFetchCall from '../utils/useFetchCall';
import {GET_ALL_ELECTIONS_ENDPOINT} from '../utils/Endpoints';

/*Assumption : for now an election is simply a json file with the following field
    - electionName: string
    - Format: []byte -> it stores the election questions 
    - electionStatus : number
    - collectivePublicKey :
    - electionID : string
*/
/*Disclaimer : 
Currently the Format parameter of an election is always a []string
called Candidates
 */

const Election = () => {

    const [context, ] = useContext(LanguageContext);
    const token = sessionStorage.getItem('token');
    const request = {
        method: 'POST',
        body: JSON.stringify({'Token': token})
    }
    const [data, loading, error] = useFetchCall(GET_ALL_ELECTIONS_ENDPOINT, request);


    /*Show all the elections retrieved if any */
    const showElection = ()=>{
        return (
            <div>
                {data.AllElectionsInfo.length > 0 ? (<div className='election-box'>
                <div className='click-info'>{Translations[context].clickElection}</div>
                    <div className = 'election-table-wrapper'>
                        <ElectionTable elections={data.AllElectionsInfo} />
                    </div>   
                </div>):<div className = 'no-election'>{Translations[context].noElection}</div>}
            </div>
        )
    }

  return (
    <div className='election-wrapper'>
        {Translations[context].listElection}
    {!loading?
        (showElection() )   
        : 
        (error===null?<p className='loading'>{Translations[context].loading} </p>:<div className='error-retrieving'>{Translations[context].errorRetrievingElection}</div>)}
    </div>
  );
}

export default Election;



