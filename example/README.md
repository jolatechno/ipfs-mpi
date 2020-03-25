# ipfs-mpi example

Example mpi handler

This example needs to be added to ipfs :

```
ipfs add -r example_interpretor
```

Wich will give an output ressembling :

```
added QmQq3yhJK8ZhyYBLS3REyUKk6YMJUoVQQio5eb4vLtyLYB example_interpreter/0.0.0/init.py
added QmVUiNgWfuvwUQsK5KvR5HFrxjyMfkXbfBDzZFyfVL1aNd example_interpreter/0.0.0/run.py
added QmRvCpa9BXQbzPNNbGqzY7A4k6D7Vc328bgKjpUhoFLWrZ example_interpreter/0.0.0
added QmVcGWrBqFSiicBVBdr6cD4aeivV9hKvVfDUCrkGeKezCz example_interpreter
 290 B / 290 B [=====================================================================================================================================] 100.00%
```

Where the line corresponding to the whole directory is `added QmVcGWrBqFSiicBVBdr6cD4aeivV9hKvVfDUCrkGeKezCz example_interpreter`, so the ipfs-store that you should use will be `QmVcGWrBqFSiicBVBdr6cD4aeivV9hKvVfDUCrkGeKezCz/`
