import {React,useContext, useState, useEffect} from 'react';
import './Ballot.css';
import {Translations} from '../language/Translations';
import {LanguageContext} from '../language/LanguageContext';
import useElection from '../utils/useElection';
import usePostCall from '../utils/usePostCall';
import {CAST_BALLOT_ENDPOINT} from '../utils/Endpoints';
import {encryptVote} from './VoteEncrypt';
import Modal from '../modal/Modal';
import {OPEN} from '../utils/StatusNumber';
import {Link} from 'react-router-dom';
import kyber from "@dedis/kyber";
import PropTypes from 'prop-types';


const Ballot = (props) => {//props.location.data = id of the election  
    const [context,] = useContext(LanguageContext);
    const token = sessionStorage.getItem('token');
    const {loading, title,candidates,electionID,status,pubKey} = useElection(props.location.data, token)
    const [choice, setChoice] = useState('');
    const [userErrors, setUserErrors] = useState({});
    const edCurve = kyber.curve.newCurve("edwards25519");
    const [postRequest, setPostRequest] = useState(null);
    const [postError, setPostError] = useState(null);
    const {postData} = usePostCall(setPostError);
    const [showModal, setShowModal] = useState(false);
    const [modalText, setModalText] = useState(Translations[context].voteSuccess);

    useEffect(()=>{
        if(postRequest !== null){
            setPostError(null);
            postData(CAST_BALLOT_ENDPOINT, postRequest, setShowModal);
        }
    }, [postRequest])

    useEffect(()=> {
        if(postError !== null){
            if(postError.includes('ECONNREFUSED')){
                setModalText(Translations[context].errorServerDown);
            } else {
                setModalText(Translations[context].voteFailure);
            }           
        } else {
            setModalText(Translations[context].voteSuccess);
        }
    }, [postError])

    const handleCheck = e =>{
        setChoice(e.target.value);
    }

    const hexToBytes = (hex) => {
        for (var bytes = [], c = 0; c < hex.length; c += 2)
        bytes.push(parseInt(hex.substr(c, 2), 16));
        return new Uint8Array(bytes);
    }

    const createBallot = (K,C)=>{
        let ballot = {};
        let vote = JSON.stringify({'K': Array.from(K), 'C':Array.from(C)});
        ballot['ElectionID'] = electionID; 
        ballot['UserId'] = sessionStorage.getItem('id');       
        ballot['Ballot'] = [...Buffer.from(vote)];
        ballot['Token'] = token;
        return ballot;
    }

    const sendBallot = async() => {
        const [K,C] = encryptVote(choice, Buffer.from(hexToBytes(pubKey).buffer), edCurve);

        //sending the ballot to evoting server
        let ballot = createBallot(K,C);
        let newRequest = {
            method: 'POST',
            body: JSON.stringify(ballot)
        }
        setPostRequest(newRequest);

    }

    const handleClick = () => {
        let errors = {};
        if(choice === ''){
            errors['noCandidate'] = Translations[context].noCandidate
            setUserErrors(errors);
            return;
        }
        sendBallot();
        setUserErrors({});
    }

    const electionClosedDisplay = () =>{
        return <div> {Translations[context].voteImpossible}</div>
    }

    const possibleChoice = (candidate) => {
        return (
                <div className='checkbox-full'>
                    <input 
                    type='checkbox'
                    key={candidate}
                    className = 'checkbox-candidate'
                    value = {candidate}
                    checked = {(choice === candidate)} //only one checkbox can be selected
                    onChange = {handleCheck}
                    />
                    <label className='checkbox-label'>
                        {candidate}
                    </label>
                </div>
        )}

    const ballotDisplay = () => {
        return (
            <div><h3 className = 'ballot-title'>{title}</h3>
            <div className='checkbox-text'>{Translations[context].pickCandidate}</div>
            {candidates !== null && candidates.length !== 0 ?
            candidates.map(candidate => (possibleChoice(candidate))) : <p>Default</p>}
            {candidates !== null? <div><div className='cast-ballot-error'>{userErrors.noCandidate}</div>
                <button className='cast-ballot-btn' onClick={handleClick}>{Translations[context].castVote}</button></div> : null}</div>
        )
    }

    return (    
        <div className = 'ballot-wrapper'>
            <Modal showModal={showModal} setShowModal={setShowModal} textModal = {modalText} buttonRightText={Translations[context].close} />
            {loading? (<p className='loading'>{Translations[context].loading}</p>)
                :(<div> {status === OPEN? ballotDisplay():electionClosedDisplay()}
                    <Link to='/vote'>
                        <button className='back-btn'>{Translations[context].back}</button>
                    </Link>
                </div>)}
        </div>   
    )
}

Ballot.propTypes = {
    location : PropTypes.any, 
}

export default Ballot;