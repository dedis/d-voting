/*global , describe, it, expect, beforeEach, jest, global*/
/*eslint no-undef: "error"*/

import React from 'react';
import Enzyme,{mount, render} from 'enzyme';
import {LanguageContext} from '../language/LanguageContext';
import Adapter from '@wojtekmaj/enzyme-adapter-react-17';
import { act } from "react-dom/test-utils";
Enzyme.configure({ adapter: new Adapter() });
import Action  from '../election-status/Action';
import Status from '../election-status/Status';
import { MemoryRouter } from 'react-router-dom'


describe('ChangeAction when status initialize with 1 (ongoing)', ()=> {   
    let wrapper;
    beforeEach(()=>{
        wrapper = render(<LanguageContext.Provider value = {['en',]}><Action status={1} electionID={1}/></LanguageContext.Provider>)
    })
    it('renders', () =>{
        expect(wrapper).not.toBeNull();
    });
    it('shows button to close election', () =>{
        expect(wrapper.find('button').text()).toContain('Close');
    });
    it('shows button to cancel election', () =>{
        expect(wrapper.find('button').text()).toContain('Close');
    });
})

describe('ChangeStatus when status initialize with 2 (closed)', ()=> {
    
    let wrapper;
    let stat;
    let setStat = jest.fn();
    beforeEach(()=>{
        stat = 2;
        wrapper = mount(<LanguageContext.Provider value = {['en',]}><Action status={stat} setStatus={setStat}electionID={1}/></LanguageContext.Provider>)
    })
    it('renders', () =>{
        expect(wrapper).not.toBeNull();
    });
    it('shows button to shuffle election', () =>{
        expect(wrapper.find('button').text()).toContain('Shuffle');
    });

    //idea: the decrypt button appears when status was changed
    it('clicks shuffle button and change status', async() =>{
        
        //mock fetch call to the api
        jest.spyOn(global, "fetch").mockImplementation(() => Promise.resolve(new Response()));
        
        await act(async() =>{
           wrapper.find('button').simulate('click');
        });
        wrapper.update();
        expect(setStat).toHaveBeenCalledTimes(1);
    });
})

describe('ChangeStatus when status initialize with 3 (ballots have been shuffled)', ()=> {
    
    let wrapper;
    let stat;
    let setStat = jest.fn();
    beforeEach(()=>{
        stat = 3;
        wrapper = mount(<LanguageContext.Provider value = {['en',]}><Action status={stat} setStatus={setStat}electionID={1}/></LanguageContext.Provider>)
    })
    it('renders', () =>{
        expect(wrapper).not.toBeNull();
    });
    it('shows button to decrypt election', () =>{
        expect(wrapper.find('button').text()).toContain('Decrypt');
    });

    it('clicks shuffle button and change status', async() =>{
        
        //mock fetch call to the api
        jest.spyOn(global, "fetch").mockImplementation(() => Promise.resolve(new Response()));
        
        await act(async() =>{
           wrapper.find('button').simulate('click');
        });
        wrapper.update();
        expect(setStat).toHaveBeenCalledTimes(1);
    });
})


describe('ChangeStatus when status initialize with 5 (result available)', ()=> {
    
    let wrapper;
    let setStat = jest.fn();
    let setResultAvailable = jest.fn();
    beforeEach(()=>{
        wrapper = mount(<LanguageContext.Provider value = {['en',]}><MemoryRouter><Action status={5} setStatus = {setStat} electionID={1} setResultAvailable={setResultAvailable} /></MemoryRouter></LanguageContext.Provider>)
    })
    it('renders', () =>{
        expect(wrapper).not.toBeNull();
    });
    it('shows "results available" button', () =>{
        expect(wrapper.find('button').text()).toContain('See results');
    });
})

describe('ChangeStatus when status initialize with 6 (election canceled)', ()=> {
    
    let wrapper;
    let setStat = jest.fn();
    let setResultAvailable = jest.fn();
    beforeEach(()=>{
        wrapper = mount(<LanguageContext.Provider value = {['en',]}><Status status={6} setStatus = {setStat} electionID={1} setResultAvailable={setResultAvailable} /></LanguageContext.Provider>)
    })
    it('renders', () =>{
        expect(wrapper).not.toBeNull();
    });
    it('shows "canceled" text', () =>{
        expect(wrapper.text().includes('Canceled')).toBe(true);
    });
})
