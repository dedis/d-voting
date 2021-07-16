import {React, useContext} from 'react';
import '../../App.css';
import './NavBar.css';
import logo from '../../assets/logoWithoutText.png';
import {LanguageContext} from '../language/LanguageContext';
import {Translations} from '../language/Translations';
import {NavLink} from 'react-router-dom';

const NavBar = () => {

    const [lanContext, setLanContext] = useContext(LanguageContext);
    const activeStyle = {
        
        textShadow:'0 0.015em #177368,0 -0.015em #177368,0.01em 0 #177368,-0.01em 0 #177368',
    };

  return ( 
    <div className='nav-links'>
        <ul>
            <NavLink exact to='/'>
                <li className='nav-logo'>
                    <img src={logo} alt='small logo'></img>
                </li>
            </NavLink>
            
            <NavLink  className='nodeco' exact to='/'  activeStyle={activeStyle}>
                <li title={Translations[lanContext].navBarHome}>{Translations[lanContext].navBarHome}</li>
            </NavLink>
            <NavLink title={Translations[lanContext].navBarCreate} className='nodeco' to='/create-election'   activeStyle={activeStyle}>
                <li title={Translations[lanContext].navBarCreate}>{Translations[lanContext].navBarCreate}</li>
            </NavLink>
            <NavLink title={Translations[lanContext].navBarCreate} className='nodeco' to='/elections'  activeStyle={activeStyle}>
                <li title={Translations[lanContext].navBarStatus}>{Translations[lanContext].navBarStatus}</li>
            </NavLink>
            <NavLink className='nodeco' to='/vote'  activeStyle={activeStyle}>
                <li title={Translations[lanContext].navBarVote}>{Translations[lanContext].navBarVote}</li>
            </NavLink>
            <NavLink className='nodeco' to='/results'  activeStyle={activeStyle}>
                <li title={Translations[lanContext].navBarResult}>{Translations[lanContext].navBarResult}</li>
            </NavLink>
            <NavLink className='nodeco' to='/about'  activeStyle={activeStyle}>
                <li title={Translations[lanContext].navBarAbout}>{Translations[lanContext].navBarAbout}</li>
            </NavLink>   
            <a className='nodeco'>
                <li className='last'>
                <select value={lanContext} onChange={(e)=>setLanContext(e.target.value)}>
                        <option value='en'>en</option>
                        <option value='fr'>fr</option>
                    </select>
                </li> 
            </a>         
        </ul>
    </div>
  );
}


export default NavBar;
