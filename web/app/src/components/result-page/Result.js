import React from 'react';
import './Result.css';
import {useState} from 'react';
import LinearProgress from '@material-ui/core/LinearProgress';
import Typography from '@material-ui/core/Typography';
import DownloadResult from '../election-status/DownloadResult';
import PropTypes from 'prop-types';

/*functional component that counts the ballots and display
 the result as total percentage of the votes */
const Result = ({resultData, candidates}) =>{
    const [dataToDownload, setDataToDownload] = useState(null);
    const countBallots = (result) => {
        let resultMap = {};
        for(var i = 0; i< candidates.length;i++){
            resultMap[candidates[i]] = 0;
        }
        for(var j = 0; j< result.length;j++){
           resultMap[result[j]['Vote']]  = resultMap[result[j]['Vote']] +1;
        }
        return resultMap;
    }

    const displayPercentage = (result) => {
        let resultMap = countBallots(result);
        const sortedResultMap =Object.fromEntries(Object.entries(resultMap).sort(function([,a],[,b]){return b-a}));    
        if(dataToDownload === null){
            setDataToDownload(sortedResultMap);
        }
        return Object.entries(sortedResultMap).map(([k, val])=>{
            let percentage = (val/result.length * 100);
            return (<div key = {k}>
                        <div className='progress-box'>
                            <span className='progress-box-candidate-name'>{k} :</span>
                            <div className='progress-box-in'>                   
                                <LinearProgress variant='determinate' className='progress-bar' value={percentage} />
                            </div>
                            <span className='progress-box-label'>
                                <Typography variant='body2' className='progress-label'>{percentage.toFixed(2)}%</Typography>
                            </span>
                        </div>
                    </div>)
        })
    }

    return(
        <span>
            <div className='result-title'>Result of the election:</div>
            {displayPercentage(resultData)}
            <div className = 'number-votes'>Total number of votes : {resultData.length}</div>
            <DownloadResult resultData={dataToDownload}></DownloadResult>
        </span>
    )
}
Result.propTypes = {
   resultData : PropTypes.any,
   candidates : PropTypes.array.isRequired, 
}

export default Result;