#!/bin/bash

provider="myPassword@123"
email="joao25@example.com"

if [[ ! -z "$1" ]]; then
  if [ "$1" == "google" ]; then
    provider="google-auth"
  fi
fi

if [[ ! -z "$2" ]]; then
    email="rafael.tomelin25@gmail.com"
fi

# curl -X GET http://localhost:8080/registry/tenant/teste \
#   -H "Content-Type: application/json" 

curl -X GET http://localhost:8080/registry/tenant/my-tenant \
  -H "Content-Type: application/json" 

curl -X GET http://localhost:8080/registry/user/joao25@example.com \
  -H "Content-Type: application/json" 

# curl -X GET http://localhost:8080/registry/user/invalid@example.com \
#   -H "Content-Type: application/json" 

# curl -X POST http://localhost:8080/api/registry-account \
# curl -X POST http://localhost:8080/registry/request \
#   -H "Content-Type: application/json" \
#   -d "{
#     \"name\": \"João Silva\",
#     \"email\": \"${email}\",
#     \"tenant\": \"my-tenant\",
#     \"password\": \"${provider}\"
#   }"
