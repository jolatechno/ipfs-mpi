# libp2p-mpi

####  *__WARNING : Development in progress, might contain bug, Please download releases to avoid bug__*

[![](https://img.shields.io/badge/project-IPFS-blue.svg?style=flat-square)](https://ipfs.io/)
[![](https://img.shields.io/badge/freenode-%23ipfs-blue.svg?style=flat-square)](http://webchat.freenode.net/?channels=%23ipfs)
[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

Message Passing Interface computing using libp2p as a back-bone to build computational pool.

## How to build an interpreter directory ?

This example needs to be added to ipfs :

```
ipfs add -r example_interpretor
```

Wich will give an output ressembling :

```
added QmYjH4sibLpCSNCVbHXhfCZY78Ckj2aDDMWFwup7NGJ3JD interpreter/dummy/init.py
added Qmb5jKmyFQFDceBXLCkjdfQbJrJ5fQ6KLGJHESTJ8E5mZo interpreter/dummy/run.py
added QmNYxkata4t7Pp4xvYyaQWEtZKNfUMQ9i1WUQ5p43cvbzq interpreter/echo/__pycache__/utils.cpython-37.pyc
added QmYjH4sibLpCSNCVbHXhfCZY78Ckj2aDDMWFwup7NGJ3JD interpreter/echo/init.py
added QmURFGfqZor5j5MY3XQMMyhJWpqzfWV7ZDpC2FvrDipuLU interpreter/echo/run.py
added QmXHjtA4485CKau2sr3nsjwmBfEuB7PajDY3tkf6oM6WTK interpreter/echo/utils.py
added QmYtMkhCXQr25NWdSK5cMBmv67ZsjDFPPHkJfksnD887tE interpreter/multi_echo/__pycache__/utils.cpython-37.pyc
added QmYjH4sibLpCSNCVbHXhfCZY78Ckj2aDDMWFwup7NGJ3JD interpreter/multi_echo/init.py
added QmWNp1x7CoT9gKP6ekFALhmYNNHyUq8DjbtDHKz42nCnKD interpreter/multi_echo/run.py
added QmXHjtA4485CKau2sr3nsjwmBfEuB7PajDY3tkf6oM6WTK interpreter/multi_echo/utils.py
added QmYHjsqkopRxojbRreAsG6EDC2sfWc9okQ8X2B4hcRHYtj interpreter/round_echo/__pycache__/utils.cpython-37.pyc
added QmYjH4sibLpCSNCVbHXhfCZY78Ckj2aDDMWFwup7NGJ3JD interpreter/round_echo/init.py
added QmZtWPvv8sQWXpPvdbNvjyXcRbmcnqDVRo7gfwohqYKFLp interpreter/round_echo/run.py
added QmXHjtA4485CKau2sr3nsjwmBfEuB7PajDY3tkf6oM6WTK interpreter/round_echo/utils.py
added QmUQT4c4btHFZGgcrSzzxXTstFWSe3eS32YhVCJYGqtut7 interpreter/dummy
added QmbTf6m59E2EfQBGoqJ3vHYcaoGFgBPAskS6UDyTjGpBoV interpreter/echo/__pycache__
added QmbU4m7qVx9hF4NSH1bgLyyhS1HEhoE72fVKqC6ZrCfRBU interpreter/echo
added QmeQfhcDropaiaTYHMSoabpzssckt9aQfyuYkm6xSdo1U2 interpreter/multi_echo/__pycache__
added QmW8ETYBnHeR82JeLBdXWDbHX7oZipG8zA3ojovsmzrvMK interpreter/multi_echo
added QmU7BWPUfBbPKf4Ex4dkNSi8soDM1hRvF7mLfpmTqrPJ7N interpreter/round_echo/__pycache__
added QmVkY4hE9yWRx6otwpTkR9PHXJmGsx5VnYcC3R6BK3Zaiy interpreter/round_echo
added QmR7CHExUVmzAnSw6NyTCgGt8cAALxC8pyyDDHiEgU7uxE interpreter
 7.89 KiB / 7.89 KiB [====================================================================================================================================================================================] 100.00%
 ```

Where the line corresponding to the whole directory is `added QmR7CHExUVmzAnSw6NyTCgGt8cAALxC8pyyDDHiEgU7uxE interpreter`, so the ipfs-store that you should use will be `QmR7CHExUVmzAnSw6NyTCgGt8cAALxC8pyyDDHiEgU7uxE/`

## How to build ?

```
go build
```

## Getting started

You should first launch the ipfs daemon with `ipfs daemon` wich will output :

```
API server listening on /ip4/127.0.0.1/tcp/5001
```

Where `/ip4/127.0.0.1/tcp/5001` is the url of the ipfs daemon.

You can then run the libp2p-mpi dameon using :

```
./libp2p-mpi -ipfs-api  /ip4/127.0.0.1/tcp/5001 -ipfs-store SomeIpfsDirectory/
```

For example :

```
./libp2p-mpi -ipfs-api  /ip4/127.0.0.1/tcp/5001 -ipfs-store Qmb5pxxiBDKiX9zZT3uPHeXYdAQ2keyNyk7QjzvbuAPkSe/
```

### Commands

#### List

You can list all installed interpreters using :

```
List
```

#### Start

You can start a interpreter using :

```
Start file n args...
```

with `file` being the name of the interpreter, `n` the size the number of peer to connect to and `args` a list of argument to pass to the interpreter

#### Add

You can add interpreters using :

```
Add interpreters...
```

with `interpreters` being a list of interpreter names.

#### Del

You can delete interpreters using :

```
Del interpreters...
```

with `interpreters` being a list of interpreter names.

#### exit

You can close the interface using `exit`.

## Usage

See [example](./example) for info on how to design an interpretor for ipfs-mpi.

### WARNING : Development in progress, might contain bug

# ToDo

__main changes:__

- [x] implement basic message passing using libp2p
- [x] handle fault (like unexpected peer hangup)
- [ ] automatically delete unused file

__'cosmetic' changes:__

- [x] write an example

__long-term goals:__

- [ ] implement a _filecoin-like_ cryptocurrency rewarding computation
