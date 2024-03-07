# Redpanda Wasm Transform

To rebuild the Wasm file for Redpanda, one needs to grab Redpanda's CLI `rpk`, instructions are here: https://docs.redpanda.com/current/get-started/rpk-install/

Once RPK is installed run the following commands:

```shell
rpk transform init data-transform
cd data-transform
rpk transform build
```

Now there will be a `data-transform.wasm` file in the current directory that can be checked in here.
