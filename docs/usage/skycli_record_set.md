## skycli record set

Set attributes on a record

### Synopsis


Set attributes on a record

```
skycli record set <record_id> <key=value> [<key=value> ...]
```

### Options

```
  -d, --basedir="": Base path for locating files to be uploaded
  -i, --no-warn-complex[=false]: Ignore complex values conversion warnings and convert automatically.
      --skip-asset[=false]: Do not upload assets
```

### Options inherited from parent commands

```
      --access_token="": Access token
      --api_key="": API Key
      --config="": Config file location. Default is $HOME/.skycli/config.toml
      --endpoint="": Endpoint address
  -p, --private[=false]: Database. Default is Public.
```

### SEE ALSO
* [skycli record](skycli_record.md)	 - Modify records in database

###### Auto generated by spf13/cobra on 23-Mar-2016
