#!/usr/bin/env bash

EXISTING_CONTAINER=$(docker container  ls -a | grep moneybotdb | wc -l)
if [ $EXISTING_CONTAINER -ne 1 ]
then
  echo "Container 'moneybotdb' not found. Creating a new one."
  docker run --name moneybotdb -p 5432:5432 -e POSTGRES_USER=moneybotuser -e POSTGRES_PASSWORD=moneybotpassword -e POSTGRES_DB=moneybotdb -d postgres
else
  docker start moneybotdb
fi