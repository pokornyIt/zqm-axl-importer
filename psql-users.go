package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
)

const (
	tempTableUserDevice = "CREATE TABLE axl_data.axl_users_tmp AS " +
		"SELECT " +
		"(j.v ->> 'user_pkid')::varchar as user_pkid, " +
		"(j.v ->> 'device_pkid')::varchar as device_pkid, " +
		"(j.v ->> 'line_pkid')::varchar as line_pkid, " +
		"(j.v ->> 'firstname')::varchar as first_name, " +
		"(j.v ->> 'middlename')::varchar as middle_name, " +
		"(j.v ->> 'lastname')::varchar as last_name, " +
		"(j.v ->> 'userid')::varchar as user_id, " +
		"(j.v ->> 'department')::varchar as department, " +
		"(j.v ->> 'status')::int as status, " +
		"(j.v ->> 'islocaluser')::bool as is_local_user, " +
		"(j.v ->> 'uccx')::bool as has_uccx, " +
		"(j.v ->> 'directoryuri')::varchar as directory_uri, " +
		"(j.v ->> 'mailid')::varchar as mail_id, " +
		"(j.v ->> 'devicename')::varchar as device_name, " +
		"(j.v ->> 'devicedescrition')::varchar as device_description, " +
		"(j.v ->> 'dnorpattern')::varchar as line_number, " +
		"(j.v ->> 'alertingnameascii')::varchar as line_alerting_name, " +
		"(j.v ->> 'line_description')::varchar as line_description, " +
		"(j.v ->> 'cluster_name')::varchar as cluster_name " +
		"FROM json_array_elements($1::json) j(v); "
	tempTableLoginUser = "CREATE TABLE axl_data.axl_login_users_tmp AS " +
		"SELECT " +
		"(j.v ->> 'user_pkid')::varchar as user_pkid, " +
		"(j.v ->> 'firstname')::varchar as first_name, " +
		"(j.v ->> 'middlename')::varchar as middle_name, " +
		"(j.v ->> 'lastname')::varchar as last_name, " +
		"(j.v ->> 'userid')::varchar as user_id, " +
		"(j.v ->> 'department')::varchar as department, " +
		"(j.v ->> 'status')::int as status, " +
		"(j.v ->> 'islocaluser')::bool as is_local_user, " +
		"(j.v ->> 'uccx')::bool as has_uccx, " +
		"(j.v ->> 'directoryuri')::varchar as directory_uri, " +
		"(j.v ->> 'mailid')::varchar as mail_id, " +
		"(j.v ->> 'cluster_name')::varchar as cluster_name " +
		"FROM json_array_elements($1::json) j(v); "
	processTempTableUserDevice = "SELECT axl_data.axl_update_users($1::varchar, $2::text)"
	processTempTableLoginUser  = "SELECT axl_data.axl_update_login_users($1::varchar, $2::text)"
	processQmUpdate            = "SELECT * from axl_data.axl_update_qm($1::varchar, $2::varchar)"
	processCallUpdateByDevice  = "SELECT * from axl_update_couples_by_device($1::int, $2::bool)"
	processCallUpdateByLine    = "SELECT * from axl_update_couples_by_line($1::int, $2::bool)"
)

func connectDb() (conn *pgx.Conn, err error) {
	s := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", config.Zqm.DbServer, 5432, config.Zqm.DbUser, "callrec")
	log.WithField("conn", s).Debugf("Connection [%s]", s)
	cfg, err := pgx.ParseConfig(fmt.Sprintf("%s password=%s", s, config.Zqm.DbPassword))
	if err == nil {
		//cfg.Logger = logrusadapter.NewLogger(log.New())
		//cfg.LogLevel = pgx.LogLevelTrace
		conn, err = pgx.ConnectConfig(context.Background(), cfg)
	}
	return conn, err
}

func connectRunUserDeviceFunc(conn *pgx.Conn, user []UserDeviceLine) (err error) {
	d, err := json.Marshal(user)
	if err != nil {
		log.WithField("error", err.Error()).Errorf("Problem convert source data to JSON string")
		return err
	}
	return connectAndUpdateAxlTables(conn, processTempTableUserDevice, tempTableUserDevice, string(d))
}

func connectRunLoginUserFunc(conn *pgx.Conn, user []LoginUser) (err error) {
	d, err := json.Marshal(user)
	if err != nil {
		log.WithField("error", err.Error()).Errorf("Problem convert source data to JSON string")
		return err
	}
	return connectAndUpdateAxlTables(conn, processTempTableLoginUser, tempTableLoginUser, string(d))
}

func connectAndUpdateAxlTables(conn *pgx.Conn, sql string, tempTableName string, jsonString string) (err error) {
	_, err = conn.Exec(context.Background(), sql, tempTableName, jsonString)
	if err != nil {
		log.WithField("error", err.Error()).WithFields(log.Fields{"command": sql, "table": tempTableName}).Errorf("Process AXL DB data update")
	} else {
		log.Info("Success update AXL source table")
	}
	return err
}

func connectUpdateQm(conn *pgx.Conn) (err error) {
	var msg, data string
	log.WithFields(log.Fields{"command": processQmUpdate, "role": config.Processing.DefaultRoleName,
		"team": config.Processing.DefaultTeamName}).Debug("Process QM DB data update")
	rows, err := conn.Query(context.Background(), processQmUpdate, config.Processing.DefaultTeamName, config.Processing.DefaultRoleName)
	if err != nil {
		log.WithFields(log.Fields{"error": err.Error(), "command": processQmUpdate, "role": config.Processing.DefaultRoleName,
			"team": config.Processing.DefaultTeamName}).Errorf("Process QM DB data update")
	} else {
		log.WithField("command", "connectUpdateQm").Info("Success update QM users")
		defer rows.Close()
		for rows.Next() {
			err = rows.Scan(&msg, &data)
			if err == nil {
				if msg == "ADD" {
					log.WithFields(log.Fields{"operation": msg, "user": data}).Infof("Add new user to QM")
				} else if msg == "UPDATE" {
					log.WithFields(log.Fields{"operation": msg, "user": data}).Infof("Update existing user to QM")
				} else if msg == "DELETE" {
					log.WithFields(log.Fields{"operation": msg, "user": data}).Infof("Mark user deleted and rename it")
				} else if msg == "PROBLEM" {
					log.WithFields(log.Fields{"operation": msg, "user": data}).Error("Problem update/insert users")
				} else if msg == "PARAM" {
					log.WithFields(log.Fields{"operation": msg, "parameter": data}).Info("Use parameters for insert")
				} else {
					log.WithFields(log.Fields{"operation": msg, "message": data}).Debug("Undefined process message")
				}
			} else {
				log.WithField("error", err).Error("problem read row data")
				break
			}
		}
	}
	return err
}

func connectUpdateCalls(conn *pgx.Conn, sql string) (err error) {
	var msg, data string

	log.WithFields(log.Fields{"command": sql,
		"hours_back": config.Processing.HoursBack, "set_direction": config.Processing.SetDirection}).Debug("Process DB couple data update")
	rows, err := conn.Query(context.Background(), sql, config.Processing.HoursBack, config.Processing.SetDirection)
	if err != nil {
		log.WithFields(log.Fields{"error": err.Error(), "command": sql,
			"hours_back": config.Processing.HoursBack}).Errorf("Process DB call data update")
	} else {
		log.WithField("command", "connectUpdateCalls").Info("Success update call data")
		defer rows.Close()
		for rows.Next() {
			err = rows.Scan(&msg, &data)
			if err == nil {
				if msg == "PREPARE" {
					log.WithFields(log.Fields{"process": msg, "records": data}).Infof("Prepare couples to processing")
				} else if msg == "UPDATE" {
					log.WithFields(log.Fields{"process": msg, "records": data}).Infof("Updated couples")
				} else if msg == "LAST" {
					log.WithFields(log.Fields{"process": msg, "last_ts": data}).Infof("Stored last update timestamp from couples")
				} else {
					log.WithFields(log.Fields{"process": msg, "message": data}).Debug("Undefined process message")
				}
			} else {
				log.WithField("error", err).Error("problem read row data")
				break
			}
		}
	}
	return err
}
