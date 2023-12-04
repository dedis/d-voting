# Automated Tests w/ Playwright

## Setup

Run

```
npx playwright install-deps --dry-run
```

to be shown the dependencies that you need to install on your machine (requires `root` access).

You need to have the D-Voting application and DELA network running. You also need to make sure
that the environment variables from the D-Voting application are set (e.g. `FRONT_END_URL`) in
the shell you'll be executing the tests in.

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
