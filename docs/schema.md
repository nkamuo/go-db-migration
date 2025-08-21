Just so you know, you can generate a schema from any database you want and use it to very other databases.

For instance,
```JSON
{
    "DB": {
        "default": {
            "host": "localhost",
            "port": 5432,
            "username": "DBA",
            "password": "p2programs",
            "database": "FIRST_STSX20_DB"
        },
        "connections":[
            {
                "name": "james",
                "database": "stsservoy_preview_test_01"
            },
            {
                "name": "01",
                "database": "stsservoy_preview_test_04"
            }
        ]
        
    }
}
```

Then generate the schema from the default config
```SH
.\migrator schema export -f json -o schema.01.json
```

Or use other connection

```SH
.\migrator schema export -f json -o schema.01.json -c bzi
```

Then use this schema to validate or compare other dbs.
```SH
 .\migrator schema compare -c 01 -s .\schema.01.json
```

```SH
 .\migrator validate all -c connection_name -s .\schema.01.json
 ```