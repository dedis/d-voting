/*global , describe, it, expect*/
/*eslint no-undef: "error"*/

import React from 'react';
import Login from '../login/Login';
import renderer from 'react-test-renderer';
import {LanguageContext} from '../language/LanguageContext'

describe('Login', ()=> {
    it('should render the Login Component correctly', () => {  
        const component = renderer.create(<LanguageContext.Provider value = {['en',]}><Login /></LanguageContext.Provider>);
        let tree = component.toJSON();
    expect(tree).toMatchSnapshot();
    });
})