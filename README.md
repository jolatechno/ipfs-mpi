# libp2p-mpi

####  *__WARNING: Development in progress, might contain bugs, Please download releases to avoid bugs__*

[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

Message Passing Interface computing using libp2p as a back-bone to build computational pool.

## *__Development of version 3 in progress__*

See [libp2p-core/v3](https://github.com/jolatechno/libp2p-mpi-core/tree/master/v3) for more informations.

This version will use ethereum smart contracts to handle message sharing which will remove the need for peer reseting since messages are held by thr contract on the blockChain.

Fault tollerence will be handled by the possibility of multiple node calculating independently the same results.

## How does it work ?

The basic algorithms behind libp2p-mpi are described in [libp2p-mpi-core/](https://github.com/jolatechno/libp2p-mpi-core/).

## How to build an interpreter directory ?

This example needs to be added to ipfs :

```
ipfs add -r example_interpreters
```

Which will give an output resembling :

```
added QmWnZhpMzobocK3LqYokq6WNjsnNoszJcgBwjbABwmYMHS example_interpreters/echo/__pycache__/utils.cpython-37.pyc
added QmYjH4sibLpCSNCVbHXhfCZY78Ckj2aDDMWFwup7NGJ3JD example_interpreters/echo/init.py
added QmTaujxGjagaps15tSBWrFU3aHHucBGtMuxr1mgiQjzrMf example_interpreters/echo/run.py

...

added QmWjkXnR9Qc1UwFcgb3HDx8mrfzFjb3ixZMqQedDxkMqET example_interpreters/round_echo
added QmbK4D76sLDvRn2qKVbtqn95qBW7CvS1nXa6F2ynyn5xaF example_interpreters
 17.18 KiB / 17.18 KiB [==================================================================================================================================================================================] 100.00%
 ```

Where the line corresponding to the whole directory is `added QmbK4D76sLDvRn2qKVbtqn95qBW7CvS1nXa6F2ynyn5xaF example_interpreters`, so the ipfs-store that you should use will be `QmbK4D76sLDvRn2qKVbtqn95qBW7CvS1nXa6F2ynyn5xaF/`

## How to build ?

```
go build
```

## Getting started

### _Starting_

You should first launch the ipfs daemon with `ipfs daemon` which will output :

```
API server listening on /ip4/127.0.0.1/tcp/5001
```

Where `/ip4/127.0.0.1/tcp/5001` is the url of the ipfs daemon (note that `/ip4/127.0.0.1/tcp/5001` is the default url for both ipfs and libp2p-mpi so you probably won't have to specify it).

You can then run the libp2p-mpi dameon using :

```
./libp2p-mpi -ipfs-api  /ip4/127.0.0.1/tcp/5001 -ipfs-store SomeIpfsDirectory/
```

For example :

```
./libp2p-mpi -ipfs-api  /ip4/127.0.0.1/tcp/5001 -ipfs-store Qmb5pxxiBDKiX9zZT3uPHeXYdAQ2keyNyk7QjzvbuAPkSe/
```

You can use `./libp2p-mpi -h` for more info on config options.

### _Commands_

#### list

You can list all installed interpreters using :

```
list
```

#### start

You can start a interpreter using :

```
start file n args...
```

with `file` being the name of the interpreter, `n` the size the number of peers to connect to and `args` a list of argument to pass to the interpreter

#### add

You can add interpreters using :

```
add interpreters...
```

with `interpreters` being a list of interpreter names.

#### del

You can delete interpreters using :

```
del interpreters...
```

with `interpreters` being a list of interpreter names.

#### exit

You can cleanly exit the interface using `exit` (using `ctrl+c` will have exactly the same effect).

## Usage

See [example_interpreters](./example_interpreters) for examples on how to design an interpreter for libp2p-mpi, and feel free to use the [utils.py](./example_interpreters/echo/utils.py) file to simplify interactions with libp2p-mpi.

Note that on download of your interpreter the `init.py` file will be called, and if it returns an error it will remove the interpreter and mark it as `failed` without further consequence, which make it ok to check for requirements in the `init.py` file and return an error if requirements are not fulfilled.

# To-do

__main changes:__

- [x] implement basic message passing using libp2p
- [x] handle fault (like unexpected peer hangup)
- [ ] automatically delete unused file

__'cosmetic' changes:__

- [x] write an example

__long-term goal:__

- [ ] implement a _filecoin-like_ cryptocurrency rewarding computation
