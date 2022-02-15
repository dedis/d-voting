import React from 'react';
import About from '../About';
import renderer from 'react-test-renderer';

describe('About', () => {
  it('should render the About Component correctly in English', () => {
    const component = renderer.create(<About />);
    let tree = component.toJSON();
    expect(tree).toMatchSnapshot();
  });

  it('About component renders its text', () => {
    const component = renderer.create(<About />);
    expect(component.root.findByProps({ className: 'about-text' }).children);
  });
});
