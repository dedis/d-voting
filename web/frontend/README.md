# Automated Tests w/ Playwright

## Setup

To install Playwright run

```
npx install playwright
```

which will install the module and its dependencies. In
order to execute the environment necessary for the tests,
additional system dependencies need to be installed.

Run

```
npx playwright install-deps --dry-run
```

to be shown the dependencies that you need to install on your machine (requires `root` access).

You need to have the D-Voting application and DELA network running. You also need to make sure
that the environment variables from the D-Voting application are set:

- `FRONT_END_URL` must be set to your locally running instance
- `REACT_APP_DEV_LOGIN` must be set to `true`

in the shell you'll be executing the tests in.

## Run tests

Run

```
FRONT_END_URL=<local instance URL> REACT_APP_DEV_LOGIN=true npx playwright test
```

to run the tests. This will open a window in your browser w/ the test results.

To run interactive tests, run

```
FRONT_END_URL=<local instance URL> REACT_APP_DEV_LOGIN=true npx playwright test --ui
```

this will open an user interface where you can interactively run and evaluate tests.
