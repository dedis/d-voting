# Automated Tests w/ Playwright

## Setup

To install Playwright run

```
npm ci
npm install playwright
```

which will install the module and its dependencies. In
order to execute the environment necessary for the tests,
additional system dependencies need to be installed.

Run

```
npx playwright install-deps --dry-run
```

to be shown the dependencies that you need to install on your machine (requires `root` access).

Your local frontend must be accessible at `http://127.0.0.1:3000`.

## Run tests

Run

```
npx playwright test
```

to run the tests. This will open a window in your browser w/ the test results.

To run interactive tests, run

```
npx playwright test --ui
```

this will open an user interface where you can interactively run and evaluate tests.

## Update HAR files

To update the HAR files, you need to make sure

* that a complete D-Voting setup is running, as the API will be called for real, and
* the `REACT_APP_SCIPER_ADMIN` of the D-Voting instance value is set to `123456` (i.e. the `SCIPER_ADMIN` value in the mocks).

You then change the `UPDATE = false` value in `tests/mocks.ts` to `UPDATE = true` and execute
the tests as usual. The tests that update the mocks will be run and the other tests will be skipped.
