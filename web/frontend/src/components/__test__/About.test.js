/*global , describe, it, expect*/
/*eslint no-undef: "error"*/

import React from 'react';
import About from '../About';
import {LanguageContext} from '../language/LanguageContext';
import renderer from 'react-test-renderer';

describe('About', () => {
  it('should render the About Component correctly in English', () => {  
      const component = renderer.create(<LanguageContext.Provider value = {['en',]}><About /></LanguageContext.Provider>);
      let tree = component.toJSON();
  expect(tree).toMatchSnapshot();
  });

  it('About component renders its text', () => {  
    const component = renderer.create(<LanguageContext.Provider value = {['en',]}><About /></LanguageContext.Provider>);
    expect(component.root.findByProps({className: "about-text"}).children);
});


});
