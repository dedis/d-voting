import {React, useContext} from 'react';
import './Home.css';
import {Translations} from '../language/Translations';
import {LanguageContext} from '../language/LanguageContext';

const Home = () => {
  const [context, ] = useContext(LanguageContext);

  return(
    <div className='home'>
      <h1>{Translations[context].homeTitle}</h1>  
      <div className='home-txt'>{Translations[context].homeText}</div>
    </div>
  );
}

export default Home;