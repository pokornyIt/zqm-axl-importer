SET client_min_messages TO WARNING;

/*
  This create under new axlUser
 */
drop table if exists axl_data.axl_users cascade;
create table axl_data.axl_users
(
    user_pkid          varchar(128),                     -- AXL pkid from enduser table
    device_pkid        varchar(128),                     -- AXL pkid from device table
    line_pkid          varchar(128),                     -- AXL pkid from numplan table
    first_name         varchar(64),
    middle_name        varchar(64),
    last_name          varchar(64),
    user_id            varchar(144),
    department         varchar(64),
    status             int       default 1     not null, -- 0= disabled user 1= enabled
    is_local_user      bool      default true  not null, -- false sync from AD
    has_uccx           bool      default false not null, -- use enabled for UCCX
    directory_uri      varchar(256),
    mail_id            varchar(256),
    device_name        varchar(130),
    device_description varchar(512),
    line_number        varchar(64),
    line_alerting_name varchar(128),
    line_description   varchar(256),
    is_deleted_on_axl  bool      default false not null, -- for hold not updated
    wbsc_id            int       default 0     not null, -- connect id from wbsc
    date_insert        timestamp default now() not null, -- date when row inserted into table
    date_updated       timestamp default now() not null  -- last update or part of used for mark is_deleted_on_axl and final delete from table
);
comment on table axl_data.axl_users is 'Main table synchronize from AXL server';

create index axl_users_user_id_device_name_index
    on axl_data.axl_users (user_id, device_name);

create index axl_users_user_id_line_number_index
    on axl_data.axl_users (user_id, line_number);

create index axl_users_user_pkid_device_pkid_line_pkid_index
    on axl_data.axl_users (user_pkid, device_pkid, line_pkid);

drop view if exists axl_data.axl_user_device_view;
create or replace view axl_data.axl_user_device_view as
select user_pkid,
       first_name,
       last_name,
       middle_name,
       user_id,
       department,
       directory_uri,
       mail_id,
       device_name,
       device_description,
       wbsc_id,
       has_uccx
from axl_data.axl_users
where status = 1
  and is_deleted_on_axl = false
group by user_pkid, first_name, last_name, middle_name, user_id, department, directory_uri, mail_id, device_name,
         device_description, wbsc_id, has_uccx;
comment on view axl_data.axl_user_device_view is 'Help view return only user/device list';

drop view if exists axl_data.axl_user_line_view;
create or replace view axl_data.axl_user_line_view as
select user_pkid,
       first_name,
       last_name,
       middle_name,
       user_id,
       department,
       directory_uri,
       mail_id,
       line_number,
       line_alerting_name,
       line_description,
       wbsc_id,
       has_uccx
from axl_data.axl_users
where status = 1
  and is_deleted_on_axl = false
group by user_pkid, first_name, last_name, middle_name, user_id, department, directory_uri, mail_id, line_number,
         line_alerting_name, line_description, wbsc_id, has_uccx;
comment on view axl_data.axl_user_line_view is 'Help view return only user/line list';

/*
  Only for validation when create functions.
  Schema of temp import table when update AXL data.
 */
drop table if exists axl_data.axl_users_tmp;
create table axl_data.axl_users_tmp
(
    user_pkid          varchar(128),                -- AXL pkid from enduser table
    device_pkid        varchar(128),                -- AXL pkid from device table
    line_pkid          varchar(128),                -- AXL pkid from numplan table
    first_name         varchar(64),
    middle_name        varchar(64),
    last_name          varchar(64),
    user_id            varchar(144),
    department         varchar(64),
    status             int  default 1     not null, -- 0= disabled user 1= enabled
    is_local_user      bool default true  not null, -- false sync from AD
    has_uccx           bool default false not null, -- is user enabled for UCCX
    directory_uri      varchar(256),
    mail_id            varchar(256),
    device_name        varchar(130),
    device_description varchar(512),
    line_number        varchar(64),
    line_alerting_name varchar(128),
    line_description   varchar(256),
    cluster_name       varchar(255)
);
drop table if exists axl_data.axl_users_tmp;

/*
 Deleted users on AXL are deleted there after not update for mor than 3 days.
 Other mark as deleted on axl
 */
create or replace function axl_data.axl_update_users(sql CHARACTER VARYING,
                                                     json_data TEXT) RETURNS INT
    LANGUAGE plpgsql AS
$$
begin
    DROP TABLE IF EXISTS axl_data.axl_users_tmp;
    EXECUTE sql USING json_data;

    update axl_data.axl_users_tmp
    set user_pkid = axl_users_tmp.cluster_name || '_' || axl_users_tmp.user_pkid
    where user_id = user_id;

    insert into axl_data.axl_users (user_pkid, device_pkid, line_pkid, first_name, middle_name, last_name, user_id,
                                    department,
                                    status, is_local_user, directory_uri, mail_id, device_name, device_description,
                                    line_number,
                                    line_alerting_name, line_description, has_uccx)
    SELECT user_pkid,
           device_pkid,
           line_pkid,
           first_name,
           middle_name,
           last_name,
           user_id,
           department,
           status,
           is_local_user,
           directory_uri,
           mail_id,
           device_name,
           device_description,
           line_number,
           line_alerting_name,
           line_description,
           has_uccx
    from axl_data.axl_users_tmp
    where (user_pkid || device_pkid || line_pkid) not in
          (select user_pkid || device_pkid || line_pkid from axl_data.axl_users);

    update axl_data.axl_users
    set first_name=t.first_name,
        middle_name=t.middle_name,
        last_name=t.last_name,
        user_id=t.user_id,
        department=t.department,
        status=t.status,
        is_local_user=t.is_local_user,
        directory_uri=t.directory_uri,
        mail_id=t.mail_id,
        device_name=t.device_name,
        device_description=t.device_description,
        line_number=t.line_number,
        line_alerting_name=t.line_alerting_name,
        line_description=t.line_description,
        is_deleted_on_axl= false,
        has_uccx=t.has_uccx,
        date_updated=now()
    from axl_data.axl_users_tmp t
    where axl_users.user_pkid = t.user_pkid
      and axl_users.device_pkid = t.device_pkid
      and axl_users.line_pkid = t.line_pkid;

    delete from axl_data.axl_users where date_updated < now()::DATE - INTERVAL '5 days';

    update axl_data.axl_users
    set is_deleted_on_axl= true
    where (user_pkid || device_pkid || line_pkid) not in
          (select user_pkid || device_pkid || line_pkid from axl_data.axl_users_tmp);

    DROP TABLE IF EXISTS axl_data.axl_users_tmp;
    return 1;
end;
$$;
comment on function axl_data.axl_update_users(sql CHARACTER VARYING, json_data TEXT) is 'Bulk data update';


/*
 Table for user with valid role in CUCM for allow login to QM
 */
DROP TABLE IF EXISTS axl_data.axl_login_users CASCADE;
CREATE TABLE axl_data.axl_login_users
(
    user_pkid         varchar(128),                     -- AXL pkid from enduser table
    first_name        varchar(64),
    middle_name       varchar(64),
    last_name         varchar(64),
    user_id           varchar(144),
    department        varchar(64),
    status            int       default 1     not null, -- 0= disabled user 1= enabled
    is_local_user     bool      default true  not null, -- false sync from AD
    has_uccx          bool      default false not null, -- is user enabled for UCCX
    directory_uri     varchar(256),
    mail_id           varchar(256),
    is_deleted_on_axl bool      default false not null, -- for hold not updated
    wbsc_id           int       default 0     not null, -- connect id from wbsc
    date_insert       timestamp default now() not null, -- date when row inserted into table
    date_updated      timestamp default now() not null  -- last update or part of used for mark is_deleted_on_axl and final delete from table
);
comment on table axl_data.axl_login_users is 'User synchronize from AXL server with associate role';

create index axl_login_users_user_id_index
    on axl_data.axl_login_users (user_id);

/*
 Deleted users on AXL are deleted there after not update for more than 3 days.
 Other mark as deleted on axl
 */
drop table if exists axl_data.axl_login_users_tmp;
create table axl_data.axl_login_users_tmp
(
    user_pkid     varchar(128),                -- AXL pkid from enduser table
    first_name    varchar(64),
    middle_name   varchar(64),
    last_name     varchar(64),
    user_id       varchar(144),
    department    varchar(64),
    status        int  default 1     not null, -- 0= disabled user 1= enabled
    is_local_user bool default true  not null, -- false sync from AD
    has_uccx      bool default false not null, -- is user enabled for UCCX
    directory_uri varchar(256),
    mail_id       varchar(256),
    cluster_name  varchar(255)
);
drop table if exists axl_data.axl_login_users_tmp;

create or replace function axl_data.axl_update_login_users(sql CHARACTER VARYING,
                                                           json_data TEXT) RETURNS INT
    LANGUAGE plpgsql AS
$$
begin
    DROP TABLE IF EXISTS axl_data.axl_login_users_tmp;
    EXECUTE sql USING json_data;

    update axl_data.axl_login_users_tmp
    set user_pkid = axl_login_users_tmp.cluster_name || '_' || axl_login_users_tmp.user_pkid
    where user_id = user_id;

    insert into axl_data.axl_login_users (user_pkid, first_name, middle_name, last_name, user_id,
                                          department,
                                          status, is_local_user, directory_uri, mail_id, has_uccx)
    SELECT user_pkid,
           first_name,
           middle_name,
           last_name,
           user_id,
           department,
           status,
           is_local_user,
           directory_uri,
           mail_id,
           has_uccx
    from axl_data.axl_login_users_tmp
    where user_pkid not in
          (select user_pkid from axl_data.axl_login_users);

    update axl_data.axl_login_users
    set first_name=t.first_name,
        middle_name=t.middle_name,
        last_name=t.last_name,
        user_id=t.user_id,
        department=t.department,
        status=t.status,
        is_local_user=t.is_local_user,
        directory_uri=t.directory_uri,
        mail_id=t.mail_id,
        is_deleted_on_axl= false,
        has_uccx= t.has_uccx,
        date_updated=now()
    from axl_data.axl_login_users_tmp t
    where axl_login_users.user_pkid = t.user_pkid;

    delete from axl_data.axl_login_users where date_updated < now()::DATE - INTERVAL '5 days';

    update axl_data.axl_login_users
    set is_deleted_on_axl= true
    where user_pkid not in
          (select user_pkid from axl_data.axl_login_users_tmp);

    DROP TABLE IF EXISTS axl_data.axl_login_users_tmp;
    return 1;
end;
$$;
comment on function axl_data.axl_update_login_users(sql CHARACTER VARYING, json_data TEXT) is 'Bulk data update for login allowed users';


/*
 Help function for fix maximal string len
 */
create or replace function axl_data.fix_varchar_len(source character varying, maxlen integer) returns character varying
    language plpgsql
as
$$
BEGIN
    IF length(source) <= maxlen THEN RETURN source; END IF;
    RETURN left(source, maxlen);
END;
$$;
comment on function axl_data.fix_varchar_len(varchar, integer) is 'Change max len of varchar to defined value';


/*
 Update QM users based on AXL data
 */
create or replace function axl_data.axl_update_qm(default_team varchar(50), default_role varchar(255))
    RETURNS table
            (
                operation varchar,
                user_name varchar
            )
    LANGUAGE plpgsql
AS
$$
declare
    var_r record;
begin
    -- temp table with messages
    drop table if exists message;
    create TEMP table message
    (
        operation varchar,
        user_name varchar
    );

    -- create tem table
    drop table if exists tmp_axl_users;
    create TEMP table tmp_axl_users
    (
        name               varchar(64),                     --first_name
        surname            varchar(64),                     -- last_name
        login              varchar(144) unique,             -- user_id
        database           int          default 4,
        sync               bool         default false,
        has_uccx           bool         default false,
        status             varchar(50)  default 'INACTIVE', -- ACTIVE,INACTIVE,DELETED, alternative BLOCKED
        phone              varchar(64)  default null,       -- line_number
        agentid            varchar(128),                    --user_pkid
        identificator_used varchar(50)  default 'EXTERNAL_AGENT_ID',
        language           int          default 1,
        company            int          default 1,
        external_id        varchar(255) default null,
        daemon             bool         default false,
        email              varchar(255)
    );

    -- user for mapping
    insert into tmp_axl_users(name, surname, login, agentid, email, has_uccx)
    select a.first_name,
           a.last_name,
           a.user_id,
           a.user_pkid,
           user_mail,
           has_uccx
    from (select user_pkid,
                 u.user_id,
                 first_name,
                 last_name,
                 is_deleted_on_axl,
                 case when mail_id is null then directory_uri else mail_id end user_mail,
                 has_uccx
          from axl_data.axl_users u
          group by user_pkid, u.user_id, first_name, last_name, is_deleted_on_axl,
                   case when mail_id is null then directory_uri else mail_id end, has_uccx) a
    where a.is_deleted_on_axl = false;

    -- user for login
    insert into tmp_axl_users(name, surname, login, agentid, email, status, has_uccx)
    select a.first_name,
           a.last_name,
           a.user_id,
           a.user_pkid,
           user_mail,
           'ACTIVE',
           has_uccx
    from (select user_pkid,
                 u.user_id,
                 first_name,
                 last_name,
                 is_deleted_on_axl,
                 case when mail_id is null then directory_uri else mail_id end user_mail,
                 has_uccx
          from axl_data.axl_login_users u
          group by user_pkid, u.user_id, first_name, last_name, is_deleted_on_axl,
                   case when mail_id is null then directory_uri else mail_id end, has_uccx) a
    where a.is_deleted_on_axl = false
      and user_pkid not in (select user_pkid
                            from axl_data.axl_users
                            where is_deleted_on_axl = false);

    -- update mapping for login too
    update tmp_axl_users
    set status='ACTIVE'
    from axl_data.axl_login_users
    where tmp_axl_users.agentid = axl_data.axl_login_users.user_pkid
      and axl_data.axl_login_users.is_deleted_on_axl = false;

    RAISE NOTICE 'Finish prepare temp users table';

    -- mark delete users from AXL
    insert into message (operation, user_name)
    select 'DELETE'::varchar, login:: varchar
    from wbsc.sc_users
    where agentid not in
          (select agentid from tmp_axl_users where has_uccx = false)
      and status <> 'DELETED'
      and database = 4
      and external_id is null;

    update wbsc.sc_users
    set status='DELETED',
        deleted_ts=now(),
        login = axl_data.fix_varchar_len(login || '_exp_' || floor(extract(epoch from now()))::text, 50)
    where agentid not in
          (select agentid from tmp_axl_users where has_uccx = false)
      and status <> 'DELETED'
      and database = 4
      and external_id is null;
    RAISE NOTICE 'Finish delete users';


    -- fix name for deleted from ccx importer
    insert into message (operation, user_name)
    select 'DELETE other'::varchar, login:: varchar
    from wbsc.sc_users
    where lower(login) in (select lower(login) from tmp_axl_users where has_uccx = false)
      and status <> 'DELETED'
      and database = 4
      and external_id is not null
      and login ~ '.*_ccx_\d+$';

    update wbsc.sc_users
    set deleted_ts=now(),
        login     = axl_data.fix_varchar_len(login || '_ccx_' || floor(extract(epoch from now()))::text, 50)
    where lower(login) in (select lower(login) from tmp_axl_users where has_uccx = false)
      and status = 'DELETED'
      and database = 4
      and external_id is not null;
    RAISE NOTICE 'Finish delete other users';


    -- insert new users and password
    insert into message (operation, user_name)
    select 'ADD'::varchar, a.login::varchar
    from (select agentid, login from tmp_axl_users u where has_uccx = false) a
    where a.agentid not in (select agentid from wbsc.sc_users where agentid is not null)
      and lower(a.login) not in (select lower(login) from wbsc.sc_users);

    insert into wbsc.sc_users (name, surname, login, database, sync, status, phone, agentid, identificator_used,
                               language, company, external_id, daemon, email)
    select name,
           surname,
           login,
           database,
           sync,
           status,
           phone,
           agentid,
           identificator_used,
           language,
           company,
           external_id,
           daemon,
           email
    from tmp_axl_users a
    where a.agentid not in (select agentid from wbsc.sc_users where agentid is not null)
      and lower(a.login) not in (select lower(login) from wbsc.sc_users)
      and has_uccx = false;

    RAISE NOTICE 'Finish insert new users';

    -- add new user to required role
    INSERT INTO wbsc.user_role (userid, roleid)
    SELECT userid, roleid
    from (
             select userid
             from wbsc.sc_users
             where agentid in (select agentid from tmp_axl_users)
               and userid not in (select userid from wbsc.user_role)) as a,
         (select roleid
          from wbsc.roles
          where name = default_role) as r;
    RAISE NOTICE 'Finish insert role for new users';

    -- add new user to default team
    INSERT INTO wbsc.user_belongsto_ccgroup (ccgroupid, userid)
    SELECT ccgroupid, userid
    from (
             select userid
             from wbsc.sc_users
             where agentid in (select agentid from tmp_axl_users)
               and userid not in (select userid from wbsc.user_belongsto_ccgroup)) as a,
         (select ccgroupid
          from wbsc.ccgroups
          where ccgroupname = default_team) as r;
    RAISE NOTICE 'Finish insert group for new users';

    -- update back wbsc_id
    update axl_data.axl_users
    set wbsc_id = s.userid
    from wbsc.sc_users s
    where wbsc_id = 0
      and s.agentid = user_pkid;

    update axl_data.axl_login_users
    set wbsc_id = s.userid
    from wbsc.sc_users s
    where wbsc_id = 0
      and s.agentid = user_pkid;

    RAISE NOTICE 'Finish update back axl_data';

    -- Update exist QM not delete record if some different
    insert into message (operation, user_name)
    select 'UPDATE'::varchar, a.login::varchar
    from (select agentid,
                 u.login,
                 name,
                 surname,
                 email,
                 status
          from tmp_axl_users u
          where has_uccx = false) a,
         wbsc.sc_users s
    where s.agentid = a.agentid
      and a.login not in (select login from wbsc.sc_users where agentid <> a.agentid)
      and (
            s.name != a.name or
            s.surname != a.surname or
            s.login != a.login or
            coalesce(s.email, '') != coalesce(a.email, '') or
            s.status != a.status
        );

    insert into message (operation, user_name)
    select 'PROBLEM'::varchar, a.login::varchar
    from (select agentid,
                 u.login,
                 name,
                 surname,
                 email,
                 status
          from tmp_axl_users u
          where has_uccx = false) a,
         wbsc.sc_users s
    where s.agentid = a.agentid
      and a.agentid in (select login from wbsc.sc_users where agentid <> a.agentid)
      and (
            s.name != a.name or
            s.surname != a.surname or
            s.login != a.login or
            coalesce(s.email, '') != coalesce(a.email, '') or
            s.status != a.status
        );

    update wbsc.sc_users
    set name=a.name,
        surname=a.surname,
        login=a.login,
        email=a.email,
        status=a.status
    from (select agentid,
                 u.login,
                 name,
                 surname,
                 email,
                 status
          from tmp_axl_users u
          where has_uccx = false) a
    where wbsc.sc_users.agentid = a.agentid
      and a.agentid not in (select login from wbsc.sc_users where agentid <> a.agentid)
      and (
            wbsc.sc_users.name != a.name or
            wbsc.sc_users.surname != a.surname or
            wbsc.sc_users.login != a.login or
            coalesce(wbsc.sc_users.email, '') != coalesce(a.email, '') or
            wbsc.sc_users.status != a.status
        );

    RAISE NOTICE 'Finish update users';
    drop table if exists tmp_axl_users;

    for var_r IN (select message.operation, message.user_name from message)
        LOOP
            operation := var_r.operation;
            user_name := var_r.user_name;
            return next;
        end loop;
    drop table message;
end;
$$;
comment on function axl_data.axl_update_qm(varchar, varchar) is 'Update QM users based on AXL data';


/*
  Create and fill table for last couple update
 */
drop table if exists axl_data.couple_last_update;
create table axl_data.couple_last_update
(
    id                    int unique,
    last_process          timestamp default now(),
    last_couple_update_ts timestamp default now() - '10 months'::INTERVAL
);
comment on table axl_data.couple_last_update is 'Hold last update ';

insert into axl_data.couple_last_update (id)
VALUES (1);

/*
  Update calls add agentid from sc_user base on device name and days back
 */
create or replace function axl_data.axl_update_couples_by_device(hours_back int, set_direction bool)
    RETURNS table
            (
                operation   varchar,
                description varchar
            )
    LANGUAGE plpgsql
AS
$$
declare
    var_r   record;
    cnt     int;
    last_ts timestamp;
begin
    -- message table
    drop table if exists couple_message;
    create TEMP table couple_message
    (
        operation   varchar,
        description varchar
    );
    insert into couple_message (operation, description)
    values ('START'::varchar, to_char(now(), 'YYYY-MM-DD HH24:MI:SS TZ'));

    -- validate last update table
    select count(1) into cnt from axl_data.couple_last_update where id = 1;
    if cnt < 1 then
        insert into axl_data.couple_last_update (id, last_couple_update_ts)
        VALUES (1, now() - 2 * hours_back * '1 hours'::INTERVAL);
    end if;

    -- temp table for hold necessary data
    drop table if exists couple_new_id_tmp;
    create temp table couple_new_id_tmp
    (
        id               integer unique,
        calling_agent    varchar(255),
        called_agent     varchar(255),
        calling_terminal varchar(255) default null,
        called_terminal  varchar(255) default null,
        couple_updated   timestamp,
        new_direction    varchar(25)
    );
    insert into couple_new_id_tmp (id, calling_agent, called_agent, couple_updated, new_direction)
    select id, callingagent, calledagent, updated_ts, direction
    from callrec.couples c
    where updated_ts >= (select last_couple_update_ts from axl_data.couple_last_update where id = 1 LIMIT 1)
      and created_ts >= now() - hours_back * '1 hours'::INTERVAL
      and (callingagent is null or calledagent is null);

    -- Add JTAPI names
    update couple_new_id_tmp
    set calling_terminal=value
    from callrec.couple_extdata
    where key = 'JTAPI_CALLING_TERMINAL_SEP'
      and id = cplid;

    update couple_new_id_tmp
    set called_terminal=value
    from callrec.couple_extdata
    where key = 'JTAPI_CALLED_TERMINAL_SEP'
      and id = cplid;

    select count(1) into cnt from couple_new_id_tmp;
    insert into couple_message (operation, description)
    values ('PREPARE', '' || cast(cnt as varchar(15)));

    -- update agents
    update couple_new_id_tmp
    set calling_agent=axl_data.axl_user_device_view.user_pkid
    from axl_data.axl_user_device_view
    where calling_terminal = axl_data.axl_user_device_view.device_name;

    update couple_new_id_tmp
    set called_agent=axl_data.axl_user_device_view.user_pkid
    from axl_data.axl_user_device_view
    where called_terminal = axl_data.axl_user_device_view.device_name;

    if set_direction then
        begin
            update couple_new_id_tmp
            set new_direction = case
                                    when calling_agent is not null and called_agent is not null then 'INTERNAL'::varchar
                                    when calling_agent is null and called_agent is not null then 'INCOMING'::varchar
                                    when calling_agent is not null and called_agent is null then 'OUTGOING'::varchar
                                    else new_direction end
            where calling_agent is not null
               or called_agent is not null;
        end;
    end if;

    select count(1)
    into cnt
    from (
             select 1
             from callrec.couples cc
                      inner join couple_new_id_tmp c
                                 on cc.id = c.id
                                     and not (coalesce(cc.calledagent, '') = coalesce(c.called_agent, '') and
                                              coalesce(cc.callingagent, '') = coalesce(c.calling_agent, '') and
                                              direction = new_direction
                                         )
         ) a;

    insert into couple_message (operation, description)
    values ('UPDATE', '' || cast(cnt as varchar(15)));

    update callrec.couples
    set calledagent=called_agent,
        callingagent=calling_agent,
        direction=new_direction
    from couple_new_id_tmp c
    where callrec.couples.id = c.id
      and not (coalesce(calledagent, '') = coalesce(c.called_agent, '') and
               coalesce(callingagent, '') = coalesce(c.calling_agent, '') and
               direction = c.new_direction
        );

    select max(couple_updated) into last_ts from couple_new_id_tmp;
    if last_ts is null then
        select now() - hours_back * '1 hours'::INTERVAL into last_ts;
    end if;
    update axl_data.couple_last_update
    set last_process=now(),
        last_couple_update_ts=last_ts
    where id = 1;

    insert into couple_message (operation, description)
    values ('LAST', '' || to_char(last_ts, 'YYYY-MM-DD HH24:MI:SS TZ'));
    drop table if exists couple_new_id_tmp;

    RAISE NOTICE 'Finish update couples';
    for var_r IN (select couple_message.operation, couple_message.description from couple_message)
        LOOP
            operation := var_r.operation;
            description := var_r.description;
            return next;
        end loop;
--     return QUERY (select couple_message.operation, couple_message.description from couple_message);
    drop table couple_message;

end;
$$;
comment on function axl_data.axl_update_couples_by_device(hours_back int, set_direction bool) is 'Update CallREC couples based on axl_user devices';


/*
  Update calls add agentid from sc_user base on line and days back
 */
create or replace function axl_data.axl_update_couples_by_line(hours_back int, set_direction bool)
    RETURNS table
            (
                operation   varchar,
                description varchar
            )
    LANGUAGE plpgsql
AS
$$
declare
    var_r   record;
    cnt     int;
    last_ts timestamp;
begin
    -- message table
    drop table if exists couple_message;
    create TEMp table couple_message
    (
        operation   varchar,
        description varchar
    );
    insert into couple_message (operation, description)
    values ('START'::varchar, to_char(now(), 'YYYY-MM-DD HH24:MI:SS TZ'));

    -- validate last update table
    select count(1) into cnt from axl_data.couple_last_update where id = 2;
    if cnt < 1 then
        insert into axl_data.couple_last_update (id, last_couple_update_ts)
        VALUES (2, now() - 2 * hours_back * '1 hours'::INTERVAL);
    end if;

    -- temp table for hold necessary data
    drop table if exists couple_new_id_tmp;
    create temp table couple_new_id_tmp
    (
        id             integer unique,
        calling_agent  varchar(255),
        called_agent   varchar(255),
        calling_dn     varchar(255) default null,
        called_dn      varchar(255) default null,
        couple_updated timestamp,
        new_direction  varchar(25)
    );
    insert into couple_new_id_tmp (id, calling_agent, called_agent, calling_dn, called_dn, couple_updated,
                                   new_direction)
    select id, callingagent, calledagent, callingnr, originalcallednr, updated_ts, direction
    from callrec.couples
    where updated_ts >= (select last_couple_update_ts from axl_data.couple_last_update where id = 2 LIMIT 1)
      and created_ts >= now() - hours_back * '1 hours'::INTERVAL
      and (callingagent is null or calledagent is null);

    select count(1) into cnt from couple_new_id_tmp;
    insert into couple_message (operation, description)
    values ('PREPARE', '' || cast(cnt as varchar(15)));

    -- add agent connection
    update couple_new_id_tmp
    set calling_agent=axl_data.axl_user_line_view.user_pkid
    from axl_data.axl_user_line_view
    where calling_dn = axl_data.axl_user_line_view.line_number;

    update couple_new_id_tmp
    set called_agent=axl_data.axl_user_line_view.user_pkid
    from axl_data.axl_user_line_view
    where called_dn = axl_data.axl_user_line_view.line_number;

    if set_direction then
        begin
            update couple_new_id_tmp
            set new_direction = case
                                    when calling_agent is not null and called_agent is not null then 'INTERNAL'::varchar
                                    when calling_agent is null and called_agent is not null then 'INCOMING'::varchar
                                    when calling_agent is not null and called_agent is null then 'OUTGOING'::varchar
                                    else new_direction end
            where calling_agent is not null
               or called_agent is not null;
        end;
    end if;


    select count(1)
    into cnt
    from (
             select 1
             from callrec.couples cc
                      inner join couple_new_id_tmp c
                                 on cc.id = c.id
                                     and not (coalesce(cc.calledagent, '') = coalesce(c.called_agent, '') and
                                              coalesce(cc.callingagent, '') = coalesce(c.calling_agent, '') and
                                              direction = new_direction
                                         )
         ) a;

    insert into couple_message (operation, description)
    values ('UPDATE', '' || cast(cnt as varchar(15)));

    update callrec.couples
    set calledagent=called_agent,
        callingagent=calling_agent,
        direction=new_direction
    from couple_new_id_tmp c
    where callrec.couples.id = c.id
      and not (coalesce(calledagent, '') = coalesce(c.called_agent, '') and
               coalesce(callingagent, '') = coalesce(c.calling_agent, '') and
               direction = new_direction
        );

    select max(couple_updated) into last_ts from couple_new_id_tmp;
    if last_ts is null then
        select now() - hours_back * '1 hours'::INTERVAL into last_ts;
    end if;
    update axl_data.couple_last_update
    set last_process=now(),
        last_couple_update_ts=last_ts
    where id = 2;

    insert into couple_message (operation, description)
    values ('LAST', '' || to_char(last_ts, 'YYYY-MM-DD HH24:MI:SS TZ'));
    drop table if exists couple_new_id_tmp;

    RAISE NOTICE 'Finish update couples';
    for var_r IN (select couple_message.operation, couple_message.description from couple_message)
        LOOP
            operation := var_r.operation;
            description := var_r.description;
            return next;
        end loop;
--     return QUERY (select couple_message.operation, couple_message.description from couple_message);
    drop table couple_message;

end;
$$;
comment on function axl_data.axl_update_couples_by_line(hours_back int, set_direction bool) is 'Update CallREC couples based on axl_user line';
