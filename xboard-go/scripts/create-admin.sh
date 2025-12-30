#!/bin/bash

# Script to create an admin user interactively

set -e

echo "=== Xboard Go - Create Admin User ==="
echo

# Check if we're running in Docker
if [ -f /.dockerenv ]; then
    DB_HOST=${DB_HOST:-mariadb}
else
    DB_HOST=${DB_HOST:-localhost}
fi

DB_PORT=${DB_PORT:-3306}
DB_USER=${DB_USER:-xboard}
DB_PASSWORD=${DB_PASSWORD:-xboard_password}
DB_NAME=${DB_NAME:-xboard_go}

# Prompt for admin details
read -p "Admin email: " ADMIN_EMAIL
read -sp "Admin password: " ADMIN_PASSWORD
echo
read -sp "Confirm password: " ADMIN_PASSWORD_CONFIRM
echo

if [ "$ADMIN_PASSWORD" != "$ADMIN_PASSWORD_CONFIRM" ]; then
    echo "Error: Passwords do not match"
    exit 1
fi

# Generate password hash using a simple Go program
HASH=$(cat <<EOF | go run -
package main
import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
    "os"
)
func main() {
    hash, _ := bcrypt.GenerateFromPassword([]byte(os.Args[1]), bcrypt.DefaultCost)
    fmt.Print(string(hash))
}
EOF
"$ADMIN_PASSWORD")

# Insert into database
mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" <<SQL
INSERT INTO users (email, password_hash, role, banned, created_at, updated_at)
VALUES (
    '$ADMIN_EMAIL',
    '$HASH',
    'admin',
    0,
    NOW(),
    NOW()
);
SQL

echo
echo "Admin user created successfully!"
echo "Email: $ADMIN_EMAIL"
echo "You can now login via the web interface or API."
