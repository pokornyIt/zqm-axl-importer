SET client_min_messages TO WARNING;

-- DROP ALL TABLES
DROP FUNCTION IF EXISTS axl_data.axl_update_qm(varchar, varchar) CASCADE;
DROP FUNCTION IF EXISTS axl_data.axl_update_couples_by_device(int) CASCADE; -- old before version 2.1
DROP FUNCTION IF EXISTS axl_data.axl_update_couples_by_device(int, bool) CASCADE;
DROP FUNCTION IF EXISTS axl_data.axl_update_couples_by_line(int) CASCADE; -- old before version 2.1
DROP FUNCTION IF EXISTS axl_data.axl_update_couples_by_line(int, bool) CASCADE;
DROP FUNCTION IF EXISTS axl_data.axl_update_login_users(varchar, text) CASCADE;
DROP FUNCTION IF EXISTS axl_data.axl_update_users(varchar, text) CASCADE;
DROP FUNCTION IF EXISTS axl_data.fix_varchar_len(varchar, integer) CASCADE;
DROP TABLE IF EXISTS axl_data.couple_last_update CASCADE;
DROP TABLE IF EXISTS axl_data.axl_login_users CASCADE;
DROP TABLE IF EXISTS axl_data.axl_users CASCADE;

-- REVOKE ACCESS TO SCHEMAS
REVOKE ALL ON SCHEMA axl_data FROM wbscgrp;
REVOKE ALL ON SCHEMA axl_data FROM callrecgrp;
REVOKE ALL ON SCHEMA axl_data FROM axlUser;

-- CLEANUP USER
DO
$$
    DECLARE
        count int;
    BEGIN
        SELECT count(*) INTO count FROM pg_roles WHERE rolname = 'axlUser';
        IF count > 0 THEN
            EXECUTE 'REVOKE CONNECT ON DATABASE "callrec" FROM axlUser';
        END IF;
    END
$$;

-- CLEANUP SCHEMA
DROP SCHEMA IF EXISTS axl_data;

-- No privileges left, now it should be possible to drop
DROP USER IF EXISTS axlUser;
