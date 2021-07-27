# ZQM AXL Importer

    Version: 1.2

### Usage
    zqm-axl-importer --config=server.json [--cli | --show | --version]   
    zqm-axl-importer -h|--help   

#####PARAMETERS  
    --config=server.json    Configuration file (JSON or YAML format)  
    --cli                   Run only once and ends  
    --show                  Show actual configuration and ends 
    --version               Show program version  
    -h                      Show help
    --help                  Show help

## DATABASE

Under postgres administrator create new schema (from file `01_createschema.sql`).
In file can change name of user and password. 

Use process file `02_createtable.sql` for create necessary table and functions.

##Configuration file

System support configuration file in JSON or YAML format. Options are same in both.
```json
{
  "axl": {
    "server": "c09-cucm-a.devlab.zoomint.com",
    "user": "ccmadmin",
    "password": "zoomadmin",
    "ignoreCertificate": true
  },
  "zqm": {
    "jtapiUser": "callrec",
    "dbServer": "pm028.pm.zoomint.com",
    "dbPort": 5432,
    "dbUser": "axluser",
    "dbPassword": "a4lUs3r."
  },
  "log": {
    "level": "TRACE",
    "fileName": "./log/data.log",
    "jsonFormat": false,
    "logProgramInfo": true
  },
  "processing": {
    "hoursBack": 24,
    "userImportHour": [
      4,
      16
    ],
    "defaultTeamName": "_CUCM_imported",
    "defaultRoleName": "Agent",
    "updateInterval": 5
  }
}
```
