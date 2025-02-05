![Dobo](_dev/docs/logo.webp?raw=true)

# dobo
A write-optimized database written in Go, inspired by Cassandra DB (LSM tree)

## Setting up a Node

Generate an async key pair & certificate:

    ```bash
    openssl req -new -newkey rsa:2048 -nodes -keyout nodekey.key -x509 -out nodekey.crt

    openssl rsa -pubout -in nodekey.key -out nodekey.pem
    ```

Create a config file called `node.toml`, and add your generated keys & cert:

```toml
[socket]
port=2480

private_key="/path/to/nodekey.key"
public_key="/path/to/nodekey.pem"
cert="/path/to/nodekey.crt"
```

Now you can run the node as such:

    ```bash
    ./obodb /path/to/node.toml
    ```
