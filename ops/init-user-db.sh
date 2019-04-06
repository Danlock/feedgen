#!/bin/bash
# Polls CockroachDB and creates user and db 

echo "Installing netcat"
apt-get update
apt-get install -y netcat
echo "Polling CRDB"
while ! nc -w 1 -z localhost 26257; do sleep 0.1; done;
echo "Initializing CRDB";
/cockroach/cockroach sql --insecure --user=root --execute="
	CREATE USER IF NOT EXISTS $COCKROACH_USER;
	CREATE DATABASE IF NOT EXISTS $COCKROACH_DATABASE;
	GRANT CREATE, SELECT, DROP, INSERT, DELETE, UPDATE ON DATABASE $COCKROACH_DATABASE TO $COCKROACH_USER;
";