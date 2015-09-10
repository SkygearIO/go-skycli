# Example

## Importing record with an asset image

    $ cat hongkong.json
    {
        "_id": "city/hongkong",
        "name": "Hong Kong",
        "image": "@images/hongkong.jpg"
    }
    $ odcli record import hongkong.json
    Found an asset in the "image" key of record "city/hongkong". Continue? (y/n) y
    $

To use stdin,

    $ echo '{ "_id": "city/hongkong", "name": "Hong Kong", "image": "@images/hongkong.jpg" }' | odcli record import -i

Alternatively,

    $ odcli record set city/hongkong image=@images/hongkong.jpg name="Hong Kong"

## Exporting record

    # mkdir cities
    # odcli record export -o cities city/hongkong city/paris city/london
    # cat cities/city-hongkong.json
    {"_id":"city/hongkong","name":"HongKong","image":"@file:city-hongkong.jpg"}
    # file cities/city-hongkong.jpg
    cities/city-hongkong.jpg: JPEG image data

Without exporting asset

    # odcli record export --skip-asset -p city/hongkong
    {
        "_id": "city/hongkong",
        "name": "Hong Kong",
        "image": "@asset:af88670f-de82-4e50-9682-3408829c0d77"
    }

To download an asset

    # odcli record get -a -o=hongkong.jpg city/hongkong image
