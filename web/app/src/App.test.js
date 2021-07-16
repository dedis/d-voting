/*global beforeEach, describe, it, expect*/
/*eslint no-undef: "error"*/
import Enzyme, {mount, shallow} from 'enzyme';
import Adapter from '@wojtekmaj/enzyme-adapter-react-17';
Enzyme.configure({ adapter: new Adapter() });
import App from './App';
import React from 'react';
import {Route} from 'react-router-dom';
import Login from './components/login/Login';
import Footer from './components/footer/Footer';


describe('App testing without being authenticated', ()=> {

  let wrapper;
  beforeEach(()=>{
      wrapper = shallow(<App />);
  })

  it('renders without crashing', () =>{
    expect(wrapper).not.toBeNull();
  });
  
  it('renders 2 <Route /> components when no token', () => {
    console.log(wrapper.debug());
    expect(wrapper.find(Route).length).toBe(2);
  })

  it('renders <Login /> when no token', ()=>{
    expect(wrapper.find(Login).length).toBe(1);
  })

  it('renders the navigation bar', () => {
    const wrapper = mount(<App />);
    const navBar = 'Home';
    expect(wrapper.contains(navBar)).toEqual(true);
  })

  it('renders the footer', () => {
    const wrapper = mount(<App />);
    expect(wrapper.find(Footer).length).toEqual(1);
  })

})

