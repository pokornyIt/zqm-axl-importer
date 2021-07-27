#!/bin/bash
# This installation script deploys the callrec-zqm-axl-importer tool
# Version 2.1.0_01
# Tested with allrec-zqm-axl-importer tool versions: 2.1.x
# 2.1.0_01 Version update:      Added Configuration option Support for Co-existance with UCCX importer 
#                               Added Configuration option Support for setting Call direction 
#                               Fixed: Tomcat parameters after re-install are missing ssl=false parameter
# 2.0.2_02 Version update:      Added initial user import 
#                               Fixed: Subtitution for mapping mode .
# 2.0.2_01 Version update:      Enhancement to support calls-user mapping method.
#				Enhancement for missing configuration of older config file
# 2.0.1_01 Version update: 	Fixed: Editing of callrec-tomcat paramaters.
#				Added versioning

VERSION="2.1.0_01"

# Installation files
SOURCE_FILES_SQL="00_cleanup.sql 01_createschema.sql 02_createtable.sql"
SOURCE_FILES_TOOL="config.json flush_sc_cache.txt jmxterm-1.0.1-uber.jar zqm-axl-importer remove_zqm_axl_importer.sh install_zqm_axl_importer.sh"
SOURCE_FILE_SERVICE="callrec-zqm-axl-importer.service"

CONFIG_FILE="config.json"

SOURCE_DIR=$(dirname "$(readlink -f "$0")")
SOURCE_SQL_DIR="$SOURCE_DIR"
TARGET_DIR="/opt/zqm-axl"
TARGET_SQL_DIR="$TARGET_DIR/SQL"
TARGET_SERVICE_DIR="/lib/systemd/system/"

# Configuration file default values
DEF_CONF_CUCM_SERVER="cucm.server"
DEF_CONF_AXL_USER="cucm.user"
DEF_CONF_AXL_PASS="secret.password"
DEF_CONF_CUCM_ACCESS_GRP="accessGroup"
DEF_CONF_JTAPI_USER="jtapi.user"
DEF_CONF_DB_HOST="localhost"
DEF_CONF_TEAM_NAME="_CUCM_imported"
DEF_CONF_ROLE_NAME="Agent"
DEF_CONF_IMPORT_HOURS="userImportHour"
DEF_CONF_MAPPING_MODE="mappingType"
DEF_CONF_DB_PORT="dbPort"
DEF_CONF_DB_USER="dbUser"
DEF_CONF_DB_PASS="dbPassword"
DEF_CONF_COEXISTUCCX="coexistCcxImporter"
DEF_CONF_SET_DIRECTION="setDirection"


BACKUP_DATE=$(date -d 'now' +%Y%m%d_%H%M%S)
BACKUP_FILE="${TARGET_DIR}/${SRC_FILE}_back${BACKUP_DATE}"

function start_installation() {
  echo
  echo "INFO The installation and configuration of callrec-zqm-axl-importer tool started."
  echo "     Installer version : ${VERSION}"
  echo "-----------------------------------"
}

function jsonValue() {
  SEP_OBJECT_START="{"
  SEP_OBJECT_END="}"
  SEP_PROPERTY=","
  SEP_PROPERTY_VALUE=":"
  SEP_ARRAY_START="["
  SEP_ARRAY_END="]"
  SEP_ARRAY_VALUE=","

  OBJECT="" PROPERTY=""
  FILENAME=""

  # get options
  PRINT_HELP="0"

  while getopts ":hf:o:p:" opt; do
    case $opt in
    h)
      PRINT_HELP="1"
      ;;
    f)
      FILENAME="$OPTARG"
      ;;
    o)
      OBJECT="$OPTARG"
      ;;
    p)
      PROPERTY="$OPTARG"
      ;;
    \?)
      echo "Unknown option $OPTARG, aborting ..."
      exit 1
      ;;
    esac
  done

  if [ ${PRINT_HELP} -eq 1 ]; then
    SCRIPT_NAME=$(basename "$0")
    echo "Simple json file parser in bash"
    echo "Usage:    ./${SCRIPT_NAME} -f <filename.json> -o <obejctname> -p <propertyname>"
    echo "Example: config.json"
    echo "---------"
    echo "{ "
    echo " \"axl\": { "
    echo "    \"server\": \"cucm.server\","
    echo "    \"user\": \"cucm.user\","
    echo "  }"
    echo "}"
    echo "---------"
    echo
    echo "Usage example 1: "
    echo " ./${SCRIPT_NAME} -f config.json -o axl -p user"
    echo "Output:"
    echo " cucm.user"
    echo "----------"
    echo
    echo "Usage example 2: "
    echo " ./${SCRIPT_NAME} -f config.json -o axl"
    echo "Output:"
    echo " \"axl\": { "
    echo "    \"server\": \"cucm.server\","
    echo "    \"user\": \"cucm.user\","
    echo "  }"
    return 0
  fi

  # Variable FILENAME is SET and it is empty -> Exit function 0
  if [ -z "${FILENAME}" ]; then
    return 0
  fi

  # Fills is the requested object. If object is empty, whole file is loaded
  if [ -z "${OBJECT}" ]; then
    OBJECT_CONTENT=$(cat ${FILENAME})
    echo "${OBJECT_CONTENT}"
    return 1
  else
    OBJECT_CONTENT=$(sed -n "/\"${OBJECT}\".*\:.*${SEP_OBJECT_START}/,/${SEP_OBJECT_END}/p" ${FILENAME})
  fi

  # We will search for the property within the object
  if [ -z "${PROPERTY}" ]; then
    PROPERTY_CONTENT="${OBJECT_CONTENT}"
  else
    # Raw read of property value. It may be type of Array
    PROPERTY_CONTENT_RAW=$(echo "${OBJECT_CONTENT}" | sed -n "/\"${PROPERTY}\".*\:.*/,/${SEP_PROPERTY}/p")
    PROPERTY_CONTENT_RAW=$(echo "${PROPERTY_CONTENT_RAW}" | sed -n -e "0,/${SEP_PROPERTY}/p")
    # check whether the returned value might be an array
    if [ $(echo "${PROPERTY_CONTENT_RAW}" | grep \\${SEP_ARRAY_START} | wc -c) -gt 1 ]; then
      IS_ARRAY=1
    else
      IS_ARRAY=0
    fi

    # for NON- ARRAY we use simple approach ... just the value
    if [ ${IS_ARRAY} -eq 0 ]; then
      PROPERTY_CONTENT=$(echo "${PROPERTY_CONTENT_RAW}" | sed -e "s|\"${PROPERTY}\".*\:||")

      # first sed removes trailing spaces before string
      # second sed removes trailing spaces after string
      # third sed removes the , separator at the ned of the string
      # fourth sed removes remaining trailing spaces at the enf if exists
      # head sorts out the last possible value in the OBJECT, which does not require to be finished by , separator
      # remaining 2 sed remove " for strings to get similar result to ARRAY output for string values
      PROPERTY_CONTENT=$(echo "${PROPERTY_CONTENT}" | sed -e 's/^[ \t]*//' -e 's/[ \t]*$//' -e 's/,$//' -e 's/[ \t]*$//' | head -n 1 | sed -e 's|^\"||' -e 's|\"$||')
    fi

    if [ ${IS_ARRAY} -eq 1 ]; then
      # it the object is array, we have to read it once again with different approach
      PROPERTY_CONTENT_RAW=$(echo "${OBJECT_CONTENT}" | sed -n "/\"${PROPERTY}\".*\:.*/,/${SEP_ARRAY_END}/p")
      PROPERTY_CONTENT_RAW=$(echo "${PROPERTY_CONTENT_RAW}" | sed -n -e "0,/${SEP_ARRAY_END}/p")
      PROPERTY_CONTENT=$(echo "${PROPERTY_CONTENT_RAW}" | sed -e "s|\"${PROPERTY}\".*\:||")

      #need to check whether array stores strings or values. Checking " character
      if [ $(echo ${PROPERTY_CONTENT} | grep "\"" | wc -c) -gt 0 ]; then
        COLUMN_SEP="\""
        AWK_BY_SPACE="0"
      else
        COLUMN_SEP=" "
        AWK_BY_SPACE="1"
      fi

      # each string starts and ends with " ... this is used a s a separator for awk
      # awk print all columns as separate lines (thad why the for loop is used
      # we need to delete characters, which we wo not want to have in the list
      # first sed removes trailing spaces before string
      # second sed removes trailing spaces after string
      # third sed removes start sign of the array [
      # fourth  sed removes start sign of the array ]
      # fifth sed  removes separators of values
      # sixth sed deletes empty lines remaining

      PROPERTY_CONTENT=$(echo ${PROPERTY_CONTENT} | awk -F "${COLUMN_SEP}" '{ for (i=1;i<=NF;i++) print $i }' | sed -e 's/^[ \t]*//' -e 's/[ \t]*$//')
      PROPERTY_CONTENT=$(echo "${PROPERTY_CONTENT}" | sed -e "s|^\\${SEP_ARRAY_START}||" -e "s|^\\${SEP_ARRAY_END}||" -e "s|^\\${SEP_ARRAY_VALUE}$||" -e '/^\s*$/d')
      if [ ${AWK_BY_SPACE} -eq 1 ]; then
        PROPERTY_CONTENT=$(echo "${PROPERTY_CONTENT}" | sed -e 's|,||g')
      fi
    fi
  fi

  echo ${PROPERTY_CONTENT}
  return 2
}

function stop_running_instance() {
  TOOL_STATUS=$(systemctl is-active callrec-zqm-axl-importer)
  if [ "${TOOL_STATUS}" == "active" ]; then
    echo "     INFO     Stopping existing instance of the callrec-zqm-axl-importer tool."
    systemctl stop callrec-zqm-axl-importer
  fi
}

function check_consistency() {
  echo "Running installation consistency check"
  SOURCE_SQL_OK="1"
  SOURCE_TOOL_OK="1"
  SOURCE_SERVICE_OK="1"
  for SRC_FILE in $(echo "${SOURCE_FILES_SQL}"); do
    if [ $(find ${SOURCE_SQL_DIR} -name ${SRC_FILE} | wc -l) -eq 0 ]; then
      SOURCE_SQL_OK="0"
      echo "     ERROR     Missing file: ${SRC_FILE}"
    fi
  done
  for SRC_FILE in $(echo "${SOURCE_FILES_TOOL}"); do
    if [ $(find ${SOURCE_DIR} -name ${SRC_FILE} | wc -l) -eq 0 ]; then
      SOURCE_TOOL_OK="0"
      echo "    ERROR     Missing file: ${SRC_FILE}"
    fi
  done
  for SRC_FILE in $(echo "${SOURCE_FILE_SERVICE}"); do
    if [ $(find ${SOURCE_DIR} -name ${SRC_FILE} | wc -l) -eq 0 ]; then
      SOURCE_SERVICE_OK="0"
      echo "     ERROR    Missing file: ${SRC_FILE}"
    fi
  done
  if [ ${SOURCE_SQL_OK} -eq 0 ]; then
    echo "     ERROR    SQL consistency check failed. Check all of the following files are present in ${SOURCE_SQL_DIR}"
    echo "${SOURCE_FILES_SQL}"
  else
    echo "     INFO     SQL consistency - all required files found"
  fi
  if [ ${SOURCE_TOOL_OK} -eq 0 ]; then
    echo "     ERROR    TOOLS  consistency check failed. Check all of the following files are present in ${SOURCE_DIR}"
    echo "${SOURCE_FILES_TOOL}"
  else
    echo "     INFO     TOOLS consistency - all required files found"
  fi
  if [ ${SOURCE_SERVICE_OK} -eq 0 ]; then
    echo "     ERROR    SERVICE Unit  consistency check failed. Check all of the following files are present in ${SOURCE_DIR}"
    echo "${SOURCE_FILE_SERVICE}"
  else
    echo "     INFO     SERVICE unit consistency - all required files found"
  fi
  CONSISTENCY_RESULT=$((SOURCE_SQL_OK * SOURCE_TOOL_OK * SOURCE_SERVICE_OK))
  if [ ${CONSISTENCY_RESULT} -eq 0 ]; then
    echo "     ERROR    Installation source files are not complete. Check the list of missing files."
    exit 0
  fi
}

function convert_to_unix() {
  for SRC_FILE in $(find ${SOURCE_DIR} -type f); do
    if [ $(file ${SRC_FILE} | grep ASCII | grep CRLF | wc -l) -gt 0 ]; then
      dos2unix -q -ascii ${SRC_FILE}
    fi
  done
}

function copy_files() {
  echo "     INFO     Creating target directory ${TARGET_DIR}"
  mkdir -p ${TARGET_SQL_DIR}

  for SRC_FILE in $(echo "${SOURCE_FILES_SQL}"); do
    echo "     INFO     Copying ${SRC_FILE}"
    yes | cp ${SOURCE_SQL_DIR}/${SRC_FILE} ${TARGET_SQL_DIR}
  done

  for SRC_FILE in $(echo "${SOURCE_FILES_TOOL}"); do
    echo "     INFO     Copying ${SRC_FILE}"
    if [ "${SRC_FILE}" == "${CONFIG_FILE}" ]; then
      if [ -f ${TARGET_DIR}/${CONFIG_FILE} ]; then
        BACKUP_DATE=$(date -d 'now' +%Y%m%d_%H%M%S)
        BACKUP_FILE="${TARGET_DIR}/${CONFIG_FILE}_back${BACKUP_DATE}"
        cp ${TARGET_DIR}/${CONFIG_FILE} ${BACKUP_FILE}
        echo "     INFO     Configuration file already exists. File was backed up to: ${BACKUP_FILE}"
        cp ${CONFIG_FILE_TEMP} ${TARGET_DIR}/${CONFIG_FILE}
      else
        cp ${CONFIG_FILE_TEMP} ${TARGET_DIR}/${CONFIG_FILE}
      fi
    else
      yes | cp ${SOURCE_DIR}/${SRC_FILE} ${TARGET_DIR}
    fi
  done
  chmod +x ${TARGET_DIR}/zqm-axl-importer
  echo "     INFO     Copy of files is finished"
}

function enable_service() {
  echo "     INFO     Preparing callrec-zqm-axl-importer.service"
  for SRC_FILE in $(echo "${SOURCE_FILE_SERVICE}"); do
    echo "     INFO     Copying ${SRC_FILE}"
    yes | cp ${SOURCE_DIR}/${SRC_FILE} ${TARGET_SERVICE_DIR}
  done

  systemctl daemon-reload
  echo "     INFO     Enabling callrec-zqm-axl-importer.service"
  systemctl enable callrec-zqm-axl-importer.service
  echo -n "     INFO     Staring callrec-zqm-axl-importer.service"
  systemctl restart callrec-zqm-axl-importer.service
  for i in $(seq 1 5); do
    sleep 1
    echo -n ". "
  done
  echo
  echo "---------------------------------------"
  systemctl status callrec-zqm-axl-importer.service
  echo "---------------------------------------"
  echo
}

function read_config() {
  #	OLD_CONFIG=${BACKUP_FILE}
  #	OLD_CONFIG="${SOURCE_DIR}/${CONFIG_FILE}"
  OLD_CONFIG="${TARGET_DIR}/${CONFIG_FILE}"
  echo "Checking old configuration."
  if [ -z "${OLD_CONFIG}" ] || [ ! -f "${OLD_CONFIG}" ]; then
    echo "     INFO    Old configuration file not found."
    echo "     INFO    Reading default configuration file: ${SOURCE_DIR}/${CONFIG_FILE}. "
    OLD_CONFIG="${SOURCE_DIR}/${CONFIG_FILE}"
    OLD_CONFIG_FOUND="0"
  else
    OLD_CONFIG_FOUND="1"
  fi
  CUCM_SERVER=$(jsonValue -f ${OLD_CONFIG} -o axl -p server)
  AXL_USER=$(jsonValue -f ${OLD_CONFIG} -o axl -p user)
  AXL_PASS=$(jsonValue -f ${OLD_CONFIG} -o axl -p password)
  CUCM_ACCESS_GRP=$(jsonValue -f ${OLD_CONFIG} -o axl -p accessGroup) 
  JTAPI_USER=$(jsonValue -f ${OLD_CONFIG} -o zqm -p jtapiUser)
  DB_HOST=$(jsonValue -f ${OLD_CONFIG} -o zqm -p dbServer)
  DB_PORT=$(jsonValue -f ${OLD_CONFIG} -o zqm -p dbPort)
  DB_USER=$(jsonValue -f ${OLD_CONFIG} -o zqm -p dbUser)
  DB_PASS=$(jsonValue -f ${OLD_CONFIG} -o zqm -p dbPassword)
  IMPORT_USERS_HOUR=$(jsonValue -f ${OLD_CONFIG} -o processing -p userImportHour)
  IMPORT_TEAM_NAME=$(jsonValue -f ${OLD_CONFIG} -o processing -p defaultTeamName)
  IMPORT_ROLE_NAME=$(jsonValue -f ${OLD_CONFIG} -o processing -p defaultRoleName)
  MAPPING_MODE=$(jsonValue -f ${OLD_CONFIG} -o processing -p mappingType)
  SET_DIRECTION=$(jsonValue -f ${OLD_CONFIG} -o processing -p setDirection)
  COEXISTUCCX=$(jsonValue -f ${OLD_CONFIG} -o processing -p coexistCcxImporter)

 
  #set default if does not exists
 ##########################################################################
  
  if [ $OLD_CONFIG_FOUND -eq 1 ]; then
    echo "     INFO     Old configuration file ${OLD_CONFIG} found."
    echo "Old configuration:"
    echo "-----------------------------"
    echo "Loading users and associated devices from CUCM:"
    echo "     CUCM server: $CUCM_SERVER"
    echo "     CUCM AXL user username: $AXL_USER"
    echo "     CUCM AXL pass password: $AXL_PASS"
    echo "     CUCM Access Group for users to enable ZOOM application login : $CUCM_ACCESS_GRP"
    echo "     CUCM JTAPI users: $JTAPI_USER"
    echo "Store data in ZOOM Database:"
    echo "     ZOOM DB host: $DB_HOST"
    echo "     ZOOM DB port: $DB_PORT"
    echo "     ZOOM DB username: $DB_USER"
    echo "     ZOOM DB password: $DB_PASS"
    echo "ZOOM QM User import details:"
    echo "     Default team name: $IMPORT_TEAM_NAME"
    echo "     Default role name: $IMPORT_ROLE_NAME"
    echo "     User Import hours: $IMPORT_USERS_HOUR"
    echo "     Co-exist with UCCX import: $COEXISTUCCX" 
    echo "ZOOM QM Calls-User mapping details:"
    echo "     Calls-User mapping type (Device/Line/Both): $MAPPING_MODE"
    echo "     Add Call Diection to processed calls: $SET_DIRECTION"
    echo "-----------------------------"
  fi


}

function migrate_configuration() {
  if [ $OLD_CONFIG_FOUND -eq 0 ]; then
    return 5
  fi
  MIGRATION_ANSWER="0"
  while [ ${MIGRATION_ANSWER} -eq 0 ]; do
    read -p "Do you want to migrate the configuration [Y/N] :   " migrate_config_var
    case $migrate_config_var in
    y | Y)
      MIGRATION_ANSWER="1"
      MIGRATE_CONFIG="1"
      ;;
    n | N)
      MIGRATION_ANSWER="1"
      MIGRATE_CONFIG="0"
      ;;
    esac
  done
  if [ ${MIGRATE_CONFIG} -eq 0 ]; then
    echo "     INFO   Configuration will not be migrated. New configuration needs to be created."
    OLD_CONFIG_FOUND="0"
    return 5
  fi

 # set default values in case configuration migration is used and old value does not exist
 if [ ${MIGRATE_CONFIG} -eq 1 ]; then
   if [ -z ${MAPPING_MODE} ]; then
	MAPPING_MODE="both"
	echo "     INFO     Configuration for Call-User mapping was Empty. Setting the value to : ${MAPPING_MODE}"
   fi
   if [ -z ${COEXISTUCCX} ]; then
        COEXISTUCCX="false"
        echo "     INFO     Configuration for Co-existence with UCCX was Empty. Setting the value to : ${COEXISTUCCX}"
   fi
   if [ -z ${SET_DIRECTION} ]; then
        SET_DIRECTION="true"
        echo "     INFO     Configuration for Adding the call direction was Empty. Setting the value to : ${SET_DIRECTION}"
   fi
   OLD_CONFIG_FOUND="1"
 fi
  # function is followed by the create_config_temp function
}

function create_config_temp() {
  CONFIG_FILE_TEMP=${SOURCE_DIR}/${CONFIG_FILE}_tmp
  echo "     INFO     Creating temporary configuration file: ${CONFIG_FILE_TEMP}"
  yes | cp ${SOURCE_DIR}/${CONFIG_FILE} ${CONFIG_FILE_TEMP}
}

function new_values_for_config_temp() {
  if [ ${OLD_CONFIG_FOUND} -eq 1 ]; then
    return 6
  fi
  echo "New configuration is required to proceed the installation:"
  CONFIRM_VALUE="0"
  while [ ${CONFIRM_VALUE} -eq 0 ]; do
    echo "----------------------------------------------------------"
    read -p "CUCM publisher address to load users from: [single ip address] [${CUCM_SERVER}]   : " cucm_server_var
    read -p "CUCM AXL user username: [username] [${AXL_USER}]   : " axl_user_var
    read -p "CUCM AXL user password: [password] [${AXL_PASS}]   : " axl_pass_var
	    read -p "CUCM Access Group name: [Access Group Name]: [$CUCM_ACCESS_GRP]   : " cucm_access_grp_var
	    read -p "ZOOM JTAPI users used for call recording [space  separated list] [${JTAPI_USER}]    : " jtapi_user_var
	    read -p "QM Database IP address [single ip address] [${DB_HOST}]   : " db_host_var
	    read -p "QM default user group [group name (without spaces)] [${IMPORT_TEAM_NAME}]   : " import_team_name_var
	    read -p "QM default user role [role name (without space}]: [${IMPORT_ROLE_NAME}]   : " import_role_name_var
	    read -p "QM default user hours [space separated list 0..23 ]: [${IMPORT_USERS_HOUR}]   : " import_user_hour_var

	    MAPPING_ANSWER="0"
	    while [ ${MAPPING_ANSWER} -eq 0 ]; do
	    read -p "Which method should be used for mapping calls to users? (Device [D] / Line [L] / Both [B]) [D/L/B]: [${MAPPING_MODE}]  " mapping_mode_var

	    if [ -z "${mapping_mode_var}" ]; then
	      MAPPING_MODE="${MAPPING_MODE}"
	      if [ -z "${MAPPING_MODE}" ]; then
		  echo "User mapping method was not selected. Using \"both\"."
		  MAPPING_MODE="both"
	      fi
	      MAPPING_ANSWER="1"
	    else
	      case $mapping_mode_var in
	      d | D)
		MAPPING_ANSWER="1"
		MAPPING_MODE="device"
		;;
	      l | L)
		MAPPING_ANSWER="1"
		MAPPING_MODE="line"
		;;
	      b | B)
		MAPPING_ANSWER="1"
		MAPPING_MODE="both"
                ;;
	      esac
	    fi
	    done


	    COEXISTUCCX_ANSWER="0"
	    while [ ${COEXISTUCCX_ANSWER} -eq 0 ]; do
	    read -p "Enable UCCX co-exist mode? Users used for UCCX will not be imported. (True [T] / False [F] : [${COEXISTUCCX}]  " coexistuccx_var
	    if [ -z "${coexistuccx_var}" ]; then
	      COEXISTUCCX="${COEXISTUCCX}"
	      if [ -z "${COEXISTUCCX}" ]; then
		  echo "UCCX co-exist mode was not selected. Using \"false\"."
		  COEXISTUCCX="false"
	      fi
	      COEXISTUCCX_ANSWER="1"
	    else
	      case $coexistuccx_var in
	      t | T)
		COEXISTUCCX_ANSWER="1"
		COEXISTUCCX="true"
		;;
	      f | F)
		COEXISTUCCX_ANSWER="1"
		COEXISTUCCX="false"
                ;;
	      esac
	    fi
	    done

	    DIRECTION_ANSWER="0"
	    while [ ${DIRECTION_ANSWER} -eq 0 ]; do
	    read -p "Do you want the tool to add direction based on agent identification to calls?  (True [T] / False [F]: [${SET_DIRECTION}]  " set_direction_var

	    if [ -z "${set_direction_var}" ]; then
	      SET_DIRECTION="${SET_DIRECTION}"
	      if [ -z "${SET_DIRECTION}" ]; then
		  echo "Adding direction was not set. Using \"true\"."
		  SET_DIRECTION="true"
	      fi
	      DIRECTION_ANSWER="1"
	    else
	      case $set_direction_var in
	      t | T)
		DIRECTION_ANSWER="1"
		SET_DIRECTION="true"
		;;
	      f | F)
		DIRECTION_ANSWER="1"
		SET_DIRECTION="false"
                ;;
	      esac
	    fi
	    done






    echo "----------------------------------------------------------"

    if [ -z "${cucm_server_var}" ]; then
      CUCM_SERVER="${CUCM_SERVER}"
    else
      CUCM_SERVER="${cucm_server_var}"
    fi

    if [ -z "${axl_user_var}" ]; then
      AXL_USER="${AXL_USER}"
    else
      AXL_USER="${axl_user_var}"
    fi

    if [ -z "${axl_pass_var}" ]; then
      AXL_PASS="${AXL_PASS}"
    else
      AXL_PASS="${axl_pass_var}"
    fi

    if [ -z "${cucm_access_grp_var}" ]; then
      CUCM_ACCESS_GRP="${CUCM_ACCESS_GRP}"
    else
      CUCM_ACCESS_GRP="${cucm_access_grp_var}"
    fi

    if [ -z "${jtapi_user_var}" ]; then
      JTAPI_USER="${JTAPI_USER}"
    else
      JTAPI_USER="${jtapi_user_var}"
    fi

    if [ -z "${db_host_var}" ]; then
      DB_HOST="${DB_HOST}"
    else
      DB_HOST="${db_host_var}"
    fi

    if [ -z "${import_team_name_var}" ]; then
      IMPORT_TEAM_NAME="${IMPORT_TEAM_NAME}"
    else
      IMPORT_TEAM_NAME="${import_team_name_var}"
    fi

    if [ -z "${import_role_name_var}" ]; then
      IMPORT_ROLE_NAME="${IMPORT_ROLE_NAME}"
    else
      IMPORT_ROLE_NAME="${import_role_name_var}"
    fi

    if [ -z "${import_user_hour_var}" ]; then
      IMPORT_USERS_HOUR="${IMPORT_USERS_HOUR}"
    else
      IMPORT_USERS_HOUR="${import_user_hour_var}"
    fi

    echo "Please verify the values above. "
    PROCEED_CONFIG="0"
    while [ ${PROCEED_CONFIG} -eq 0 ]; do
      read -p "Are the configuration correct ? If Yes, the configuration will proceed.  [Y/N]  : " proceed_var
      case $proceed_var in
      y | Y)
        check_db_connection
        if [ ${DB_IS_READY} -gt 0 ]; then
          CONFIRM_VALUE="1"
          PROCEED_CONFIG="1"
        else
          CONFIRM_VALUE="0"
          PROCEED_CONFIG="1"
        fi

        ;;
      n | N)
        CONFIRM_VALUE="0"
        PROCEED_CONFIG="1"
        ;;
      esac
    done
  done
}

function update_config_temp() {
  echo "     INFO     Updating configuration"




  sed -i "s|\"${DEF_CONF_CUCM_SERVER}\"|\"${CUCM_SERVER}\"|" ${CONFIG_FILE_TEMP}
  sed -i "s|\"${DEF_CONF_AXL_USER}\"|\"${AXL_USER}\"|" ${CONFIG_FILE_TEMP}
  sed -i "s|\"${DEF_CONF_AXL_PASS}\"|\"${AXL_PASS}\"|" ${CONFIG_FILE_TEMP}
  sed -i "/\"${DEF_CONF_CUCM_ACCESS_GRP}\"/c \\\t\"${DEF_CONF_CUCM_ACCESS_GRP}\": \"${CUCM_ACCESS_GRP}\"\," ${CONFIG_FILE_TEMP}
  # JTAPI_USER list a a array
  JTAPI_USER=`echo ${JTAPI_USER} | sed -e 's|\ |\",\\\\n\\\\t\\\\t\"|g'`
  sed -i "s|\"${DEF_CONF_JTAPI_USER}\"|\"${JTAPI_USER}\"|" ${CONFIG_FILE_TEMP}
  sed -i "s|\"${DEF_CONF_DB_HOST}\"|\"${DB_HOST}\"|" ${CONFIG_FILE_TEMP}
  sed -i "s|\"${DEF_CONF_TEAM_NAME}\"|\"${IMPORT_TEAM_NAME}\"|" ${CONFIG_FILE_TEMP}
  sed -i "s|\"${DEF_CONF_ROLE_NAME}\"|\"${IMPORT_ROLE_NAME}\"|" ${CONFIG_FILE_TEMP}
  # IMPORT_USERS_HOUR is a array
  IMPORT_USERS_HOUR=`echo ${IMPORT_USERS_HOUR} | sed -e 's|\ |,\\\\n\\\\t\\\\t|g'`
  sed -i "/\"${DEF_CONF_IMPORT_HOURS}\"/,/]/c \\\t\"${DEF_CONF_IMPORT_HOURS}\": \[\n\t\t${IMPORT_USERS_HOUR}\n\t\]\," ${CONFIG_FILE_TEMP}
  sed -i "/\"${DEF_CONF_DB_PORT}\"/c \\\t\"${DEF_CONF_DB_PORT}\": ${DB_PORT}\," ${CONFIG_FILE_TEMP}
  sed -i "/\"${DEF_CONF_DB_USER}\"/c \\\t\"${DEF_CONF_DB_USER}\": \"${DB_USER}\"\," ${CONFIG_FILE_TEMP}
  sed -i "/\"${DEF_CONF_DB_PASS}\"/c \\\t\"${DEF_CONF_DB_PASS}\": \"${DB_PASS}\"\," ${CONFIG_FILE_TEMP}
  sed -i "s|\"${DEF_CONF_MAPPING_MODE}.*\"|\"${DEF_CONF_MAPPING_MODE}\": \"${MAPPING_MODE}\"|" ${CONFIG_FILE_TEMP} 
  sed -i "s|\"${DEF_CONF_COEXISTUCCX}.*e|\"${DEF_CONF_COEXISTUCCX}\": ${COEXISTUCCX}|" ${CONFIG_FILE_TEMP}
  sed -i "s|\"${DEF_CONF_SET_DIRECTION}.*e|\"${DEF_CONF_SET_DIRECTION}\": ${SET_DIRECTION}|" ${CONFIG_FILE_TEMP}

}

function check_db_connection() {
  echo "Checking connection to the DB callrec on host: ${DB_HOST}"
  if [ -z "${DB_HOST}" ]; then
    echo "     ERROR    Configuration for DB not found. Verify the configuration."
    DB_IS_READY="-1"
  fi
  DB_IS_READY=$(psql -U postgres -At -d callrec -h ${DB_HOST} -c"select count(1) from wbsc.sc_users where login='ipccimporterdaemon';" 2>/dev/null)
  if [ -z "${DB_IS_READY}" ]; then
    echo "     ERROR    Connection to the Database failed. Please check connection to database (Edit pg_hba.conf on DB server)"
    DB_IS_READY="-1"
  fi
  if [ ${DB_IS_READY} -eq 0 ]; then
    echo "     WARN     QM database is missing ipccimporterdaemon user. QM DB may not be initialized properly."
  fi
  if [ ${DB_IS_READY} -eq 1 ]; then
    echo "     INFO     QM database accepts connections. QM is initialized. "
  fi
  if [ ${DB_IS_READY} -gt -1 ]; then
    echo "     INFO     Connection to the DB is opened."
  fi
}

function cleanup_db() {
  echo
  echo "     INFO     Cleaning up the Database"
  echo "     ---------------------------------"
  psql --quiet -U postgres -d callrec -h ${DB_HOST} <${TARGET_SQL_DIR}/00_cleanup.sql
}

function create_db() {
  echo
  echo "     INFO     Creating new Database Tables."
  echo "     -------------------------------------"
  for sql_script in $(find ${TARGET_SQL_DIR} -name "*.sql" | grep -v "/00" | sort); do
    echo
    echo "         Running SQL script: ${sql_script}"
    echo "         --------------------------------"
    psql --quiet -U postgres -d callrec -h ${DB_HOST} <${sql_script}
  done
  psql --quiet -U postgres -d callrec -h ${DB_HOST} -c"ALTER USER axlUser PASSWORD '${AXL_PASS}';"
}

function is_finished() {
  echo
  echo "-------------------------------"
  echo "INFO Installation is finished."
  echo "     Database ${DB_HOST} was updated"
  echo "     Configuration can be found in ${TARGET_DIR}/${CONFIG_FILE} ."
  echo "     You can check the service status by command \"qm-services\"."
}

function edit_web_service_for_reload() {
  WEB_ENABLED=$(systemctl is-enabled callrec-tomcat)
  if [ -z ${WEB_ENABLED} ] || [ "${WEB_ENABLED}" == "disabled" ]; then
    echo "     WARN      Web application is not enabled on this server.  Manual configuration of callrec-tomcat service will be required."
    echo "               Skipping configuration of callrec-tomcat service."
    WEB_CONFIG_APPLIED="0"
    WEB_RESTART_REQUIRED="0"
    return 7
  fi

  CALLREC_TOMCAT_SERVICE="/etc/systemd/system/callrec-tomcat.service.d/override.conf"
  PARAMS_FOR_JMX=" -Dcom.sun.management.jmxremote -Dcom.sun.management.jmxremote.port=8765 -Dcom.sun.management.jmxremote.ssl=false -Dcom.sun.management.jmxremote.authenticate=false"
  JMX_PORT_NUMBER="8765"
  JMX_PORT="-Dcom.sun.management.jmxremote.port="
  JMX_PORT=`echo "${JMX_PORT}" | sed -e 's|\=|\\\=|' -e 's|\-|\\\-|'`
  #Environment="JAVA_OPTS=-Dhttps.protocols=TLSv1,TLSv1.1,TLSv1.2 -Xmx1536m -XX:+DisableExplicitGC -Djava.awt.headless=true"
  TOMCAT_JAVA_OPTS=$(systemctl cat callrec-tomcat | grep Environment=\"JAVA_OPTS= | grep -v \# | tail -n1)
  if [ $(echo ${TOMCAT_JAVA_OPTS} | grep "${JMX_PORT}" | wc -l) -gt 0 ]; then
    echo "     INFO     WEB parameters already have JMX port set:"
    echo "              ${TOMCAT_JAVA_OPTS}"
    echo "     INFO     Changing the JMX port to: ${JMX_PORT_NUMBER}"
    sed -i "s|${JMX_PORT}*\ |${JMX_PORT}${JMX_PORT_NUMBER}\ |" ${CALLREC_TOMCAT_SERVICE}
  else
    TOMCAT_JAVA_OPTS_JMX=$(echo ${TOMCAT_JAVA_OPTS} | sed -e "s|\"$|\ ${PARAMS_FOR_JMX}\"|")
    TOMCAT_JAVA_OPTS_OVER=$(grep "Environment=\"JAVA_OPTS=" ${CALLREC_TOMCAT_SERVICE} | tail -n1)
    TOMCAT_JAVA_OPTS_LINE=$(grep -n "Environment=\"JAVA_OPTS=" ${CALLREC_TOMCAT_SERVICE} | tail -n1 | sed -e 's|:.*$||')
    sed -i "s|^Environment=\"JAVA_OPTS=|\#Environment=\"JAVA_OPTS=|" ${CALLREC_TOMCAT_SERVICE}
    sed -i "${TOMCAT_JAVA_OPTS_LINE}a\\${TOMCAT_JAVA_OPTS_JMX}" ${CALLREC_TOMCAT_SERVICE}
  fi
  systemctl daemon-reload
}

function restart_web_ui() {
  WEB_ENABLED=$(systemctl is-enabled callrec-tomcat)
  if [ -z ${WEB_ENABLED} ] || [ "${WEB_ENABLED}" == "disabled" ]; then
    return 7
    WEB_CONFIG_APPLIED="0"
  fi
  WEB_CONFIG_APPLIED="1"
  WEB_RESTART_ANSWER="0"
  while [ ${WEB_RESTART_ANSWER} -eq 0 ]; do
    echo "     INFO     Restart of the WEB application is required to apply changes. Restart of WEB causes logout of all users from the WebUI."
    read -p "              Do you want to restart now ? [Y/N]   : " restart_web_var
    case $restart_web_var in
    y | Y)
      WEB_RESTART_ANSWER="1"
      WEB_RESTART_REQUIRED="1"
      ;;
    n | N)
      WEB_RESTART_ANSWER="1"
      WEB_RESTART_REQUIRED="0"
      ;;
    esac
  done
  if [ ${WEB_RESTART_REQUIRED} -eq 1 ]; then
    echo "     INFO     Restarting the WEB application. This may take some time."
    systemctl restart callrec-tomcat
  fi
}

function initial_import {
 TOOL_STATUS=$(systemctl is-active callrec-zqm-axl-importer) 
  if [ -z ${TOOL_STATUS} ] || [ "${TOOL_STATUS}" == "failed" ]; then
    return 8
  fi   

  FIRST_IMPORT_ANSWER=0
  while [ ${FIRST_IMPORT_ANSWER} -eq 0 ]; do
    echo "     INFO     Initial user import is recommended."
    read -p "              Do you want import users now ? [Y/N]   : " first_import_var
    case $first_import_var in
    y | Y)
      FIRST_IMPORT_ANSWER="1"
      FIRST_IMPORT_REQUIRED="1"
      ;;
    n | N)
      FIRST_IMPORT_ANSWER="1"
      FIRST_IMPORT_REQUIRED="0"
      ;;
    esac
  done

  if [ ${FIRST_IMPORT_REQUIRED} -eq 1 ]; then
    echo "     INFO     Running the initial import."
    echo "-----------------------------------------"
    ${TARGET_DIR}/zqm-axl-importer --cli --config=${TARGET_DIR}/${CONFIG_FILE}
    echo "-----------------------------------------"
  fi




}


function additional_info() {
  if [ ${WEB_CONFIG_APPLIED} -eq 0 ]; then
    echo
    echo "WARN   WEB configuration was not applied. Please finalize the installation manually. "
    echo "          on the web server: Open the FW port: ${JMX_PORT}"
    echo "          on the web server edit the callre-tomcat service. Extend JAVA_OPTS parameters with \" ${PARAMS_FOR_JMX} \"."

  fi
  echo
  if [ ${WEB_RESTART_REQUIRED} -eq 0 ]; then
    echo "WARN    WEB application was not restarted. To finish installation restart the web application manually by command: \" systemctl restart callrec-tomcat \""
  fi

  if [ ${FIRST_IMPORT_REQUIRED} -eq 0 ]; then
    echo "WARN    Initial User import was not performed. You can run the initial import by command: \"${TARGET_DIR}/zqm-axl-importer --cli --config=${TARGET_DIR}/${CONFIG_FILE} \" "
  fi
}

start_installation
check_consistency
convert_to_unix
read_config
migrate_configuration
create_config_temp
new_values_for_config_temp
update_config_temp
stop_running_instance
copy_files
cleanup_db
create_db
edit_web_service_for_reload
restart_web_ui
enable_service
initial_import
is_finished
additional_info
