## General

This project consists of two main services store and moneropay.
Both services have their own databases that I've run with docker. 
Store and moneropay I run in user space for ease of development.

## Instructions

Moneropay: 
    Terminal1:
        $ cd moneropay && docker compose up
    Terminal2:
        $ cd moneropay && make run
Store
    Terminal1:
        docker compose up
    Terminal2:
        cd backend && docker compose up
