/*
 Create new schema and add user for access this schema This script run under main administrator postgres user
 */
CREATE SCHEMA if not exists axl_data;

CREATE USER axlUser LOGIN PASSWORD 'a4lUs3r.' NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT NOREPLICATION CONNECTION LIMIT -1;
ALTER ROLE axlUser WITH LOGIN;

GRANT ALL ON SCHEMA axl_data TO axlUser;
ALTER DEFAULT PRIVILEGES in schema axl_data GRANT ALL on tables to axlUser;
ALTER DEFAULT PRIVILEGES in schema axl_data GRANT ALL on sequences to axlUser;
ALTER DEFAULT PRIVILEGES in schema axl_data GRANT ALL on functions to axlUser;

GRANT ALL ON SCHEMA axl_data TO callrecgrp;
ALTER DEFAULT PRIVILEGES in schema axl_data GRANT ALL on tables to callrecgrp;
ALTER DEFAULT PRIVILEGES in schema axl_data GRANT ALL on sequences to callrecgrp;
ALTER DEFAULT PRIVILEGES in schema axl_data GRANT ALL on functions to callrecgrp;

GRANT ALL ON SCHEMA axl_data TO wbscgrp;
ALTER DEFAULT PRIVILEGES in schema axl_data GRANT ALL on tables to wbscgrp;
ALTER DEFAULT PRIVILEGES in schema axl_data GRANT ALL on sequences to wbscgrp;
ALTER DEFAULT PRIVILEGES in schema axl_data GRANT ALL on functions to wbscgrp;

/*
 Allow axlUser login user wbsc schema
 */
GRANT wbscgrp TO axluser;
GRANT callrecgrp to axluser;

ALTER DATABASE callrec set search_path TO callrec,wbsc,public,axl_data;