# Automated Tests w/ Playwright

## Setup

To install Playwright run

```
npm ci
npx playwright install
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
