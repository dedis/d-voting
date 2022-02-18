import React from 'react';
import renderer from 'react-test-renderer';

import About from '../../pages/About';

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
