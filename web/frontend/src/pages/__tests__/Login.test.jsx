import React from 'react';
import Login from '../Login';
import renderer from 'react-test-renderer';

describe('Login', () => {
  it('should render the Login Component correctly', () => {
    const component = renderer.create(<Login />);
    let tree = component.toJSON();
    expect(tree).toMatchSnapshot();
  });
});
