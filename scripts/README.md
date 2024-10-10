# Scripts for running D-voting locally

The following scripts are available to configure and run D-voting locally. 
They should be called in this order:

- `run_local.sh` - sets up a complete system with 4 nodes, the db, the authentication-server,
  and the frontend.
  The script runs only the DB in a docker environment, all the rest is run directly on the machine.
  This allows for easier debugging and faster testing of the different parts, mainly the
  authentication-server, but also the frontend.
  For debugging Dela, you still need to re-run everything.
- `local_proxies.sh` needs to be run once after the `run_local.sh` script
- `local_forms.sh` creates a new form and prints its ID

Every script must be called from the root of the repository:

```bash
./scripts/run_local.sh
./scripts/local_proxies.sh
./scripts/local_forms.sh
```

The following script is only called by the other scripts:

- `local_login.sh` logs into the frontend and stores the cookie
