# skycli

## Overview

```

Usage: skycli <command> [OPTIONS]

  --config=~/.skycli Configuration file
  --endpoint=       Skygear endpoint address
  --api_key=        API Key

Commands:

    configure       Configure this program
    record          Fetch and saves record
    schema          Manage record schema

```

## Record

### record import
```

Usage: skycli record import [OPTIONS] [<path> ...]

Import records to database.

  --skip-asset          Do not import asset.
  -d, --basedir=        Base path for locating asset to be uploaded.
  -i, --no-warn-complex Ignore complex values conversion warnings.

IMPORT PATH

Record data stored in a file specified by <path> is imported to database.

If <path> is a directory, all files in the directory with a `.json` extension
is imported to database, as if each file is individually specified in the
command argument.

If <path> is not specified, the program reads record data from the stdin.

If record file contains value that points to a local file, the file
will be uploaded to be an asset, after which the asset value will be
saved to database together with record data. This is unless `--skip-asset`
is specified, which will ignore any keys with value pointing to an asset.

FILE FORMAT

When importing and exporting, specified file should contain record data in
JSON representation. Each file contains streaming multiple Json objects
with/without any delimiters. For each key-value pair in record, a same pair
exists in the top level JSON dictionary. Key that starts with underscore
is reserved by system for special attribute and would not be allowed,
except the `_id` for Record ID which is compulsory for each record.

The file format is mostly the same as the JSON format specified by Skygear
API specification, but skycli also supports a convenient shorthand for
specifying complex values such as asset, location and reference.

If the value of a key is such a complex one, begining the value with the string
`@` will signal to skycli that such value is a complex one.

The format of complex values are as follows:

  asset         @file:<pathtofile>
                @asset:<asset_id>
  location      @loc:<lat>,<lng>
  reference     @ref:<referenced_id>
  string        @str:<literal>

When specifying a complex value of asset using a path, the path is relative
to the location of the record file, rather than the current working directory.

To escape a literal string that begins with a `@`, prefix a literal string
with `@str:`.

When a value is a string and it begins with a `@`, skycli will warn to user
that the value will be converted to a relevant type. To skip this warning
and convert the string value to complex value, specify `--no-warn-complex`.


```
### record export
```

Usage: skycli record export [OPTIONS] <record_id> [<record_id> ...]

Export records from database.

  --skip-asset          Do not export asset.
  -d, --basedir=        Base path for saving files to be downloaded.
  -p, --pretty-print    Print output in a pretty format.
  -o, --output=         Path to save the output to. If not specified,
                        output is printed to stdout with newline delimiter.

OUTPUT

If output is not specified, the default is to print the result to
stdout with each record delimited by a newline character. If the output
is path to a directory, each record is saved individually in its own file.

When exporting records with assets, the asset will be downloaded
from the database before skycli write the output for the record. The output
value of the key pointing to the asset will contain complex value as mentioned
in FILE FORMAT.

When `--skip-asset` is specified, the asset will not be downloaded, and the
value will contain the Asset ID.

FILE FORMAT

See `record import`

```

### record delete

```

Usage: skycli record delete <record_id> [<record_id> ...]

Delete Records from database.

Each specified record is deleted from the database.

```
### record set
```

Usage: skycli record set <record_id> <key=value> [<key=value>...]

Set attributes on a record.

If value begins with "@", the rest of the value is treated as the 
file path to be uploaded to database as an asset and associated to the key.

To specify other complex values, see FILE FORMAT.

```
### record get
```

Usage: skycli record get [OPTIONS] <record_id> <key>

Get value of a record attribute.

  -o, --output=         Path to save the output to. If not specified,
                        output is printed to stdout.
  -a, --asset           If value to the key is an asset, download the
                        asset and output the content of the asset.


```

### record edit

```

Usage: skycli record edit [OPTIONS] [<record_type|<record_id>]

Edit a record.

  -n, --new             Do not fetch record from database before editing.

The program fetches the record from database and opens an editing
session with content of the record. When exit, the content of the file
is saved to database.

If `--new` is specified, skycli opens an editing session without fetching
a record. This is useful when creating a record.

If <record_type> is specified instead of <record_id>, a new <record_id>
is randomly generated. This also enables `--new`.

Saving an empty file aborts the action.


```
### record query
```

Usage: skycli record query [OPTIONS] <record_type>

Query records from database.

  --skip-asset          Do not export asset.
  -d, --basedir=        Base path for saving files to be downloaded.
  -p, --pretty-print    Print output in a pretty format.
  -o, --output=         Path to save the output to. If not specified,
                        output is printed to stdout with newline delimiter.

The query subcommand only supports fetching all records matching a specified
record type.

OUTPUT

See `record export`

FILE FORMAT

See `record import`

```
## Schema


### schema alter

```

Usage: skycli schema alter <record_type> add <column_name> <column_def>

Add a column to the record type.


Usage: skycli schema alter <record_type> mv <column_name> <new_column_name>

Rename a column in the record type.


Usage: skycli schema alter <record_type> rm <column_name>

Remove a column in the record type.


```
