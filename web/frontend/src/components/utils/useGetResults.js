import {GET_RESULT_ENDPOINT} from './Endpoints';

const useGetResults = () => {

    async function getResults(electionID, token, setError, setResult, setIsResultSet){
        const request = {
            method: 'POST',
            body: JSON.stringify({'ElectionID':electionID,'Token': token})
        }
        try{
            const response = await fetch(GET_RESULT_ENDPOINT,request);

            if(!response.ok){
                throw Error(response.statusTest);
            } else {
                let data = await response.json();
                setResult(data.Result);
                setIsResultSet(true);
            }
        } catch(error){
            setError(error);
        }       
    }
    return {getResults};
}

export default useGetResults;