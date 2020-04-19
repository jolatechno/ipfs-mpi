# libp2p-mpi

####  *__WARNING: Development in progress, might contain bugs, Please download releases to avoid bugs__*

[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

Message Passing Interface computing using libp2p as a back-bone to build computational pool.

## How to build an interpreter directory ?

This example needs to be added to ipfs :

```
ipfs add -r example_interpreters
```

Which will give an output resembling :

```
added QmQVixfS5Hb2uVy9jhVYogntCmTP7HD99omcL9doizhajL example_interpreters/echo/__pycache__/utils.cpython-37.pyc
added QmYjH4sibLpCSNCVbHXhfCZY78Ckj2aDDMWFwup7NGJ3JD example_interpreters/echo/init.py
added QmWR3NepbDyh7BP6tSfmxAR5Qj7tKG6tA87efVMxQ1grr6 example_interpreters/echo/run.py

...

added Qmci7BaxqTX2k5YPnt7UhoxLcE9bTUZGgA4epMBmh9ouxm example_interpreters/round_echo
added QmaYuPXKLbyo9QVJrbkN8W5TkgDV1r8SsMnEjg9fbh58BR example_interpreters
 9.33 KiB / 9.33 KiB [=========================================================================] 100.00%
 ```

Where the line corresponding to the whole directory is `added QmaYuPXKLbyo9QVJrbkN8W5TkgDV1r8SsMnEjg9fbh58BR example_interpreters`, so the ipfs-store that you should use will be `QmaYuPXKLbyo9QVJrbkN8W5TkgDV1r8SsMnEjg9fbh58BR/`

## How to build ?

```
go build
```

## Getting started

You should first launch the ipfs daemon with `ipfs daemon` which will output :

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

with `file` being the name of the interpreter, `n` the size the number of peers to connect to and `args` a list of argument to pass to the interpreter

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
