![main build](https://github.com/gasparian/money-transfers-api/actions/workflows/build.yml/badge.svg?branch=main)
![main tests](https://github.com/gasparian/money-transfers-api/actions/workflows/test.yml/badge.svg?branch=main)
# money-transfers-api
Simple API for money transfers between accounts, implemented in Go.  

### Building and running  

First, install dependencies:  
```
go get ./...
```  
Here I'm using only two dependencies:  
 - [toml](https://github.com/BurntSushi/toml) - for parsing the config file;  
 - [sqlite3](https://github.com/mattn/go-sqlite3) - as the embedded database;  

Then you can build and test the project:  
```
make && make test
```  
To build the static binary for linux (~12Mb):  
```
make build-static
```  
After a successful build, you only have to run binary to start the server:  
```
./apiserver --config-path="configs/apiserver.toml"
```  

### API Reference  

 - `GET /health`:  
   - `curl -v -X GET http://localhost:8010/health`;  
   - Returns `OK` if server is up and running;  
 - `POST /create-account`:  
   - Gets `balance` value:
     ```
     curl -v -X POST \
         -H "Content-Type: application/json" \
         --data '{"balance": 100}' \
         http://localhost:8010/create-account
   - Returns account structure filled with created `id`: 
     ```
     {
        "account_id":1,
        "balance":100
     }  
 - `POST /delete-account`:  
   - Gets `account_id`: 
     ```
     curl -v -X POST \
         -H "Content-Type: application/json" \
         --data '{"account_id": 1}' \
         http://localhost:8010/delete-account
   - Returns no payload - just 200 code if the deletion was successful;  
 - `POST /get-balance`:  
   - Gets `account_id`: 
     ```
     curl -v -X POST \
         -H "Content-Type: application/json" \
         --data '{"account_id": 2}' \
         http://localhost:8010/get-balance
   - Returns account with the current balance value:  
     ```
     {
        "account_id":2,
        "balance":100
     }  
 - `POST /deposit`:  
   - Gets `account_id` and `amount` of money to deposit: 
     ```
     curl -v -X POST \
         -H "Content-Type: application/json" \
         --data '{"to_account_id": 2, "amount": 100}' \
         http://localhost:8010/deposit
   - Returns account with the new balance value:  
     ```
     {
        "account_id":2,
        "balance":200
     }  
 - `POST /withdraw`:  
   - Gets `account_id` and `amount` of money to pull: 
     ```
     curl -v -X POST \
         -H "Content-Type: application/json" \
         --data '{"from_account_id": 2, "amount": 100}' \
         http://localhost:8010/withdraw
   - Returns account with the new balance value:  
     ```
     {
        "account_id":2,
        "balance":100
     }  
 - `POST /transfer`:  
   - Gets two `account_id` values and `amount` of money to transfer: 
     ```
     curl -v -X POST \
         -H "Content-Type: application/json" \
         --data '{"from_account_id": 1, "to_account_id": 2, "amount": 50}' \
         http://localhost:8010/transfer
   - Returns accounts with the new balance values:  
     ```
     {
        "to_account":
          {
            "account_id":2,
            "balance":150
          },
        "from_account":
          {
            "account_id":1,
            "balance":200
          },
        "transfer_id":4
     }
 - `POST /get-transfers`:  
   - Gets `account_id` and `n_days` period to query transfers log: 
     ```
     curl -v -X POST \
         -H "Content-Type: application/json" \
         --data '{"account_id": 2, "n_days": 1}' \
         http://localhost:8010/get-transfers
   - Returns list of transfers for the requested account:  
     ```
     [
       {
         "transfer_id":3,
         "timestamp":"2021-05-06T17:39:44.357Z",
         "from_account_id":2,
         "to_account_id":0,
         "amount":100
       },
       {
         "transfer_id":1,
         "timestamp":"2021-05-06T17:39:04.884Z",
         "from_account_id":1,
         "to_account_id":2,
         "amount":50
       },
       {
         "transfer_id":2,
         "timestamp":"2021-05-06T17:39:28.934Z",
         "from_account_id":0,
         "to_account_id":2,
         "amount":100
       }
     ]
