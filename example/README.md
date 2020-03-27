# ipfs-mpi example

## How to build an interpreter directory ?

This example needs to be added to ipfs :

```
ipfs add -r example_interpretor
```

Wich will give an output ressembling :

```
added QmYjH4sibLpCSNCVbHXhfCZY78Ckj2aDDMWFwup7NGJ3JD test_interpreters/dummy/0.0.0/init.py
added Qmb5jKmyFQFDceBXLCkjdfQbJrJ5fQ6KLGJHESTJ8E5mZo test_interpreters/dummy/0.0.0/run.py
added QmYjH4sibLpCSNCVbHXhfCZY78Ckj2aDDMWFwup7NGJ3JD test_interpreters/echo/0.0.0/init.py
added Qmf1ssqZWrMXaH5J4eji5ErEBPSSbiywib6xvs8gePFPLF test_interpreters/echo/0.0.0/run.py
added QmUQT4c4btHFZGgcrSzzxXTstFWSe3eS32YhVCJYGqtut7 test_interpreters/dummy/0.0.0
added QmZ3cvbBb82beSnMZV1KP8rhzoSFJJETrtWgPm8N7xWBPM test_interpreters/dummy
added QmSuFow3L4Dq99qB4ZoGHiMFhvHzZaYYi5dq2E56QyMY17 test_interpreters/echo/0.0.0
added QmTqGsyP7pXvR9Gy9ryZqNiWph5h16YKSWX1iT6xCh6HYK test_interpreters/echo
added QmZyay35BsdK49tDCJpJgtJzXmtimupEbvWgisFbnPpvyi test_interpreters
 525 B / 525 B [======================================================================================================================================] 100.00%
```

Where the line corresponding to the whole directory is `added QmZyay35BsdK49tDCJpJgtJzXmtimupEbvWgisFbnPpvyi test_interpreters`, so the ipfs-store that you should use will be `QmZyay35BsdK49tDCJpJgtJzXmtimupEbvWgisFbnPpvyi/`
