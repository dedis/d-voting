# Front-end of the DELA evoting system

This folder contains a front-end for admin and user operations.

![screen](screenshot.png)

It allows the user to create a new election, close/cancel it and also vote on
on-going elections. All the elections and their operations are saved on smart
contracts from dela.
 
# Setup

Have node.js installed. It has been tested with node version v17.5.0.

The web client has two parts:

- **frontend**, which is the website served to the browser
- **backend**, which performs the authentication and other services needed by the
  frontend.

By default, the frontend can run without using the backend. This is convenient
for development. This is made possible with the use of https://mswjs.io/, a
library that mocks requests by intercepting them. To run the frontend, go to the
`frontend/` folder, install the dependencies, and run the app:

```sh
cd frontend
npm install
npm start
```

If you want to use the backend, do the same operation in the `backend/` folder.
Then, launch the frontend with:

```sh
REACT_APP_NOMOCK=on npm start
```

# Running the tests

The unit tests can be found in `src/components/_test_` folder. They can be run
with `npm run test`.

# Automatic linting with VSCode

Be sure to configure "format on save" with eslint. The following configuration
can be placed in `.vscode/settings.json` at the root of the repos:

```json
{
    "editor.codeActionsOnSave": {
        "source.fixAll.eslint": true
    },
    "eslint.validate": ["javascript"]
}
```