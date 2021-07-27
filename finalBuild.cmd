echo Build Windows
set GOOS=windows
set GOARCH=amd64
:: -s - strip binary data from debug information
:: -w - remove DWARF table
::go install
::go build -i -ldflags="-s -w" -o "./builds/zqm-axl-importer.exe"
::go build -i -o "./builds/zqm-axl-importer.exe"

echo .
echo Build Linux amd64
set GOOS=linux
set GOARCH=amd64
go install
go build -i -ldflags="-s -w" -o "./builds/zqm-axl-importer"

set GOOS=
set GOARCH=

copy /y "callrec-zqm-axl-importer.service" "./builds/"
copy /y "config.yaml" "./builds/config.yaml"
copy /y "config.json" "./builds/config.json"
copy /y "flush_sc_cache.txt" "./builds/"
copy /y "jmxterm-1.0.1-uber.jar" "./builds/"
copy /y "DBScripts\01_createschema.sql" "./builds/"
copy /y "DBScripts\02_createtable.sql" "./builds/"
copy /y "DBScripts\00_cleanup.sql" "./builds/"
copy /y "install_zqm_axl_importer.sh" "./builds/"
copy /y "remove_zqm_axl_importer.sh" "./builds/"
copy /y "CUCM_User_import_into_QM.docx" "./builds/"
