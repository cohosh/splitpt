export TOR_PT_MANAGED_TRANSPORT_VER=1
export TOR_PT_SERVER_BINDADDR=obfs4-0.0.0.0:9090 
export TOR_PT_SERVER_TRANSPORTS=obfs4
export TOR_PT_ORPORT=127.0.0.1:8080
export TOR_PT_STATE_LOCATION=server-state/

lyrebird -enableLogging -logLevel DEBUG
