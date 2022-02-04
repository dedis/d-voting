/*custom hook that post a request to an endpoint */
const usePostCall = (setError) => {

    const postData = async(endpoint, request, setIsPosting) => {
        try{           
            const response = await fetch(endpoint,request);
            if(!response.ok){
                let err = await response.text()             
                throw Error(err);
            } else {
                setError(null);
                setIsPosting(prev => !prev);
                return true;
            }
        } catch(error){
            setError(error.message);
            setIsPosting(prev => !prev);
            return false;
        }
    }


    return {postData};
}

export default usePostCall;