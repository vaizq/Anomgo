## General

This project consists of two distinct services: Market and moneropay.
Both services have their own databases that I've run inside docker containers. 
I've run store and moneropay in user space for faster development.

## Instructions

Moneropay: 
    Terminal1: (postgresql for moneropay)
        $ cd moneropay && docker compose up
    Terminal2: (moneropay)
        $ cd moneropay && make run
Market
    Terminal1: (postresql for market)
        docker compose up
    Terminal2: (market)
        cd backend && docker compose up

// TODO: Script to build and run all with single command
