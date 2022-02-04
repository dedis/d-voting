/*global , describe, it, expect, beforeEach*/
/*eslint no-undef: "error"*/

import React from 'react';
import Election from '../election-status/Election';
import {LanguageContext} from '../language/LanguageContext';
import Enzyme, {shallow} from 'enzyme';
import Adapter from '@wojtekmaj/enzyme-adapter-react-17';
Enzyme.configure({ adapter: new Adapter() });

let mockData ={
    "AllElectionsInfo":[{"ElectionID":"00000000000000000000000000000000",
                        "Title":"test",
                        "Candidates":["a"],
                        "Status":1,
                        "Pubkey":"/23hxG/fgz99WKwQcT/xLOvavENEWbCPKL5ZNtu+IX0=",
                        "Result":[]},
                        {"ElectionID":"00000000000000000000000000000001",
                        "Title":"test2",
                        "Candidates":["e","f"],
                        "Status":1,
                        "Pubkey":"/23hxG/fgz99WKwQcT/xLOvavENEWbCPKL5ZNtu+IX0=",
                        "Result":[]}]};
let mockLoading = false;
let mockError;
jest.mock('../utils/useFetchCall', () => ({
    useFetchCall: () => [null, false, mockError]
}));

describe("renders without crashing", ()=> {

    let wrapper;
    beforeEach(()=>{
        wrapper = shallow(<LanguageContext.Provider value = {['en',]}><Election /></LanguageContext.Provider>);
    })
  
    it('renders without crashing', () =>{
      console.log(wrapper.debug());
      expect(wrapper).not.toBeNull();
    });

})


