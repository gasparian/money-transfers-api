![main build](https://github.com/gasparian/money-transfers-api/actions/workflows/build.yml/badge.svg?branch=main)
![main tests](https://github.com/gasparian/money-transfers-api/actions/workflows/test.yml/badge.svg?branch=main)
# money-transfers-api
[Revolut](https://www.revolut.com/) backend test assignment. Simple API for money transfers between accounts, implemented in Go.  

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
 Server uses `int64` numbers to represent the money, to make all calculations without computation errors.  
 To get the real value - just convert integer to float and divide the value by 100, and do everything in reverse order to convert real value to integer.  

 - `GET /health`:  
   - `curl -v -X GET http://localhost:8010/health`;  
   - Returns `OK` if server is up and running;  
 - `POST /api/v1/accounts`:  
   - Gets integer `balance` value:
     ```
     curl -v -X POST \
          -H "Content-Type: application/json" \
          --data '{"balance": 10000}' \
          http://localhost:8010/api/v1/accounts
   - Returns account structure filled with created `id`: 
     ```
     {
        "account_id":1,
     }  
 - `DELETE /api/v1/accounts`:  
   - Gets `account_id`: 
     ```
     curl -v -X DELETE -G \
          -d account_id=1 \
           http://localhost:8010/api/v1/accounts
   - Returns no payload - just 204 code if the deletion was successful;  
 - `GET /api/v1/accounts`:  
   - Gets `account_id`: 
     ```
     curl -v -X GET -G \
          -d account_id=1 \
          http://localhost:8010/api/v1/accounts
   - Returns account with the current balance value:  
     ```
     {
        "account_id":1,
        "balance":10000
     }  
 - `POST /api/v1/transfer-money`:  
   - Gets two `account_id` values and `amount` of money to transfer: 
     ```
     curl -v -X POST \
          -H "Content-Type: application/json" \
          --data '{"from_account_id": 1, "to_account_id": 2, "amount": 5000}' \
          http://localhost:8010/api/v1/transfer-money
   - Returns 204 status code if the money transfer was successful;  
 - `GET /api/v1/transactions`:  
   - Gets `account_id`, `n_days` period to query transactions log and `limit` on resulting list length:  
     ```
     curl -v -X GET -G \
          -d account_id=2 \
          -d n_last_days=1 \
          -d limit=3 \
          http://localhost:8010/api/v1/transactions
   - Returns list of transactions for the requested account:  
     ```
     [
       {
         "timestamp":"2021-05-16T08:56:36.953Z",
         "from_account_id":1,
         "to_account_id":2,
         "amount":5000
       },
       {
         "timestamp":"2021-05-16T09:09:32.396Z",
         "from_account_id":1,
         "to_account_id":2,
         "amount":1000
       },
       {
         "timestamp":"2021-05-16T09:09:34.423Z",
         "from_account_id":1,
         "to_account_id":2,
         "amount":1000
       }
     ]
