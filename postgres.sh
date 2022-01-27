#!/usr/bin/env bash

docker run --name moneybotdb -p 5432:5432 -e POSTGRES_USER=moneybotuser -e POSTGRES_PASSWORD=moneybotpassword -e POSTGRES_DB=moneybotdb -d postgres