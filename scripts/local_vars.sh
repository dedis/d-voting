export DATABASE_USERNAME=dvoting
export DATABASE_PASSWORD=postgres
export FRONTEND_URL="http://localhost:3000"
export DELA_PROXY_URL="http://localhost:2001"
export BACKEND_HOST="localhost"
export BACKEND_PORT="6000"
export SESSION_SECRET="session secret"
export REACT_APP_NOMOCK=on
# shellcheck disable=SC2155
export DB_PATH="$(pwd)/nodes/llmdb"
# The following two variables can be set to see log output from dela.
# For the generic GRPC module:
#export GRPC_GO_LOG_VERBOSITY_LEVEL=99
#export GRPC_GO_LOG_SEVERITY_LEVEL=info
# For the Dela proxy (info only):
#export PROXY_LOG=info
# For the Dela node itself (info, debug):
#export LLVL=debug
# Logging in without Gaspar and REACT_APP_SCIPER_ADMIN
export REACT_APP_DEV_LOGIN="true"
export REACT_APP_SCIPER_ADMIN=100100
export REACT_APP_VERSION=$(git describe --tags --abbrev=0)
export REACT_APP_BUILD=$(git describe --tags)
export REACT_APP_BUILD_TIME=$(date)

# uncomment this to enable TLS to test gaspar
#export HTTPS=true
# Create random voter-IDs to allow easier testing
export REACT_APP_RANDOMIZE_VOTE_ID="true"
