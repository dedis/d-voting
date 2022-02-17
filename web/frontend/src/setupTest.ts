// jest-dom adds custom jest matchers for asserting on DOM nodes.
// allows you to do things like:
// expect(element).toHaveTextContent(/react/i)
// learn more: https://github.com/testing-library/jest-dom
import '@testing-library/jest-dom';

import { dvotingproxy } from './mocks/dvotingproxy';

import Enzyme from 'enzyme';
import Adapter from '@wojtekmaj/enzyme-adapter-react-17';

Enzyme.configure({ adapter: new Adapter() });

jest.mock('react-i18next', () => ({
  // this mock makes sure any components using the translate hook can use it without a warning being shown
  useTranslation: () => {
    return {
      t: (str: string) => str,
      i18n: {
        changeLanguage: () =>
          new Promise(() => {
            /* no-op */
          }),
      },
    };
  },
}));

// Establish API mocking before all tests.
beforeAll(() => dvotingproxy.listen());

// Reset any request handlers that we may add during the tests,
// so they don't affect other tests.
afterEach(() => dvotingproxy.resetHandlers());

// Clean up after the tests are finished.
afterAll(() => dvotingproxy.close());
