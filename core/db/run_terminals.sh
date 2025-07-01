#!/bin/bash

DB_PORT_START=5433
NODE_PORT_START=8000

for i in {0..5}; do
  DB_PORT=$((DB_PORT_START + i))
  NODE_PORT=$((NODE_PORT_START + i))

  echo "🖥️ Launching node $i in new terminal | DB_PORT=$DB_PORT | PORT=$NODE_PORT"

  # Write a unique .env file
  ENV_FILE=".env.$NODE_PORT"
  cat > "$ENV_FILE" <<EOF
DB_HOST=localhost
DB_PORT=$DB_PORT
DB_USER=postgres
DB_PASS=postgres
DB_NAME=msgcertdb
BOOTSTRAP=127.0.0.1:8000
PORT=$NODE_PORT
EOF

  # Launch in new terminal window
  gnome-terminal -- bash -c "cp $ENV_FILE .env && go run main.go; exec bash"
  
  sleep 1
done
