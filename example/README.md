# ipfs-mpi example

####  *__WARNING : Development in progress, might contain bug, Please download releases to avoid bug__*

[![](https://img.shields.io/badge/project-IPFS-blue.svg?style=flat-square)](https://ipfs.io/)
[![](https://img.shields.io/badge/freenode-%23ipfs-blue.svg?style=flat-square)](http://webchat.freenode.net/?channels=%23ipfs)
[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

## How to build an interpreter directory ?

This example needs to be added to ipfs :

```
ipfs add -r example_interpretor
```

Wich will give an output ressembling :

```
added QmYjH4sibLpCSNCVbHXhfCZY78Ckj2aDDMWFwup7NGJ3JD test_interpreters/dummy/init.py
added Qmb5jKmyFQFDceBXLCkjdfQbJrJ5fQ6KLGJHESTJ8E5mZo test_interpreters/dummy/run.py
added QmYjH4sibLpCSNCVbHXhfCZY78Ckj2aDDMWFwup7NGJ3JD test_interpreters/echo/init.py
added QmQXxoEsJSnN4BYuAXscbRLcwUP9U566YWkTtERF8MqVbq test_interpreters/echo/run.py
added QmUQT4c4btHFZGgcrSzzxXTstFWSe3eS32YhVCJYGqtut7 test_interpreters/dummy
added QmPCFULBU5tXMiJ3BnqFHWHA6w5UCA1u5WUbDD5gRP4r5s test_interpreters/echo
added Qmb9mhSA2Zdh8MDDUtii1n7Ycfnu6DoaD56onv6jYfHwc7 test_interpreters
 518 B / 518 B [==============================================================================================================================] 100.00%
 ```

Where the line corresponding to the whole directory is `added Qmb9mhSA2Zdh8MDDUtii1n7Ycfnu6DoaD56onv6jYfHwc7 test_interpreters`, so the ipfs-store that you should use will be `Qmb9mhSA2Zdh8MDDUtii1n7Ycfnu6DoaD56onv6jYfHwc7/`
