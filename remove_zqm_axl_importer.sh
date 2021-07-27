#!/bin/bash
# This script removes the callrec-zqm-axl-importer tool
# Version 2.0.2_02
# Tested with allrec-zqm-axl-importer tool versions: 2.0.x

# 2.0.2_02 Version update:      Update version
# 2.0.2_01 Version update:      Update version
# 2.0.1_01 Version update:      Added versioning

trap "exit 1" TERM
export TOP_PID=$$

VERSION="2.0.2_02"
TARGET_SERVICE_DIR="/lib/systemd/system/"
TARGET_SERVICE_OVER_DIR="/etc/systemd/system/callrec-zqm-axl-importer.service.d"
SERVICE_NAME="callrec-zqm-axl-importer.service"

function start_remove() {
  echo
  echo "INFO The un-install of callrec-zqm-axl-importer tool started. "
  echo "     Remover version : ${VERSION}"
  echo "-------------------------------"
  PROCEED_REMOVE_ANSWER="0"
  while [ ${PROCEED_REMOVE_ANSWER} -eq 0 ]; do
    read -p "              Do you want to remove callrec-zqm-axl-importer tool now ? [Y/N]   : " remove_tool_var
    case $remove_tool_var in
    y | Y)
      PROCEED_REMOVE_ANSWER="1"
      WEB_RESTART_REQUIRED="1"
      REMOVE_REQUIRED="1"
      ;;
    n | N)
      PROCEED_REMOVE_ANSWER="1"
      REMOVE_REQUIRED="0"
      ;;
    esac
  done
  if [ ${REMOVE_REQUIRED} -eq 0 ]; then
    echo "     INFO     The callrec-zqm-axl-importer tool will not be removed. Exiting ....."
    kill -s TERM $TOP_PID
  fi
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

function read_install_info() {
  echo "     INFO    Checking installation."
  if [ $(systemctl list-units --full -all | grep ${SERVICE_NAME} | wc -l) -eq 0 ]; then
    echo "             Service ${SERVICE_NAME} does not exist. Exiting ....."
    kill -s TERM $TOP_PID

  fi
  TARGET_DIR=$(systemctl cat callrec-zqm-axl-importer.service | grep "^ExecStart=" | tail -n 1 | sed -e 's|/zqm-axl-importer\ .*$||')
  TARGET_DIR=$(echo ${TARGET_DIR} | sed -e 's|^.*\=||')
  CONFIG_FILE=$(systemctl cat callrec-zqm-axl-importer.service | grep "^ExecStart=" | tail -n 1 | sed -e 's|^.*--config=||')
  echo "             Installation directory:   ${TARGET_DIR}"
  echo "             Configuration file:       ${CONFIG_FILE}"
  INSTALL_DIR=$(echo ${TARGET_DIR} | sed -e 's|^.*\=||')
  TARGET_SQL_DIR="${TARGET_DIR}/SQL"
}

function read_config() {
  #       OLD_CONFIG=${BACKUP_FILE}
  #       OLD_CONFIG="${SOURCE_DIR}/${CONFIG_FILE}"
  #        OLD_CONFIG="${TARGET_DIR}/${CONFIG_FILE}"
  OLD_CONFIG="${CONFIG_FILE}"
  echo "Checking old configuration."
  if [ -z "${OLD_CONFIG}" ] || [ ! -f "${OLD_CONFIG}" ]; then
    echo "     INFO    Old configuration file not found."
    OLD_CONFIG_FOUND="0"
    return 2
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
  if [ $OLD_CONFIG_FOUND -eq 1 ]; then
    echo "     INFO     Old configuration file ${OLD_CONFIG} found."
    echo "Old configuration:"
    echo "-----------------------------"
    echo "Loading users and asscaited devices from CUCM:"
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
    echo "     Deafult role name: $IMPORT_ROLE_NAME"
    echo "     User Import hours: $IMPORT_USERS_HOUR"
    echo "-----------------------------"
  fi
}

function cleanup_db() {
  echo
  echo "     INFO     Cleaning up the Database"
  echo "     ---------------------------------"
  psql --quiet -U postgres -d callrec -h ${DB_HOST} <${TARGET_SQL_DIR}/00_cleanup.sql
}

function disable_service() {
  echo "     INFO     Disabling services"
  echo "     ---------------------------------"
  echo -n "     INFO     Stopping callrec-zqm-axl-importer.service"
  systemctl stop callrec-zqm-axl-importer.service
  for i in $(seq 1 5); do
    sleep 1
    echo -n ". "
  done
  echo
  echo "     INFO     Disabling callrec-zqm-axl-importer.service"
  systemctl disable callrec-zqm-axl-importer.service

  echo "     INFO     Removing callrec-zqm-axl-importer.service"
  rm -rf ${TARGET_SERVICE_OVER_DIR}
  rm -f ${TARGET_SERVICE_DIR}/${SERVICE_NAME}

  systemctl daemon-reload
}

function clean_install_dir() {
  echo "     INFO     Cleaning Installation directory."
  echo "     -----------------------------------------"
  if [ -z ${TARGET_DIR} ]; then
    echo "     INFO     Installation directory not found. Cleanup of directory will be skipped."
    return 4
  fi
  if [ -z ${OLD_CONFIG_FOUND} ]; then
    echo "     INFO     Configuration file not found."
    return 5
  fi
  if [ ${OLD_CONFIG_FOUND} -eq 0 ]; then
    echo "     INFO     Configuration file not found."
    return 6
  fi
  if [ ${OLD_CONFIG_FOUND} -eq 1 ]; then


    KEEP_OLD_CONFIG_ANSWER="0"
    while [ ${KEEP_OLD_CONFIG_ANSWER} -eq 0 ]; do
      read -p "              Do you want to keep configuration file:  ${OLD_CONFIG} ? [Y/N]   : " remove_oldconfig_var
      case $remove_oldconfig_var in
      y | Y)
        KEEP_OLD_CONFIG_ANSWER="1"
        KEEP_OLD_CONFIG_REQUIRED="1"
        ;;
      n | N)
        KEEP_OLD_CONFIG_ANSWER="1"
        KEEP_OLD_CONFIG_REQUIRED="0"
        ;;
      esac
    done
    cd /tmp
    if [ ${KEEP_OLD_CONFIG_REQUIRED} -eq 0 ]; then
      echo "     INFO     Removing installation directory."
      rm -rf ${INSTALL_DIR}
    else

      echo "     INFO     Removing installation files. Configuration file will be kept."
      for file in $(find ${INSTALL_DIR} -type f | grep -v ${CONFIG_FILE}$); do
        rm -f $file
      done
      rm -rf ${INSTALL_DIR}/SQL
    fi

  fi

}

function remove_web_jmx_config() {
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
  JMX_PORT=$(echo "${JMX_PORT}" | sed -e 's|\=|\\\=|' -e 's|\-|\\\-|')
  PARAMS_FOR_JMX=$(echo "${PARAMS_FOR_JMX}" | sed -e 's|\=|\\\=|' -e 's|\-|\\\-|' -e "s|\=8765|\=${JMX_PORT_NUMBER}|")
  # Environment="JAVA_OPTS=-Dhttps.protocols=TLSv1,TLSv1.1,TLSv1.2 -Xmx1536m -Xss1280k -XX:+DisableExplicitGC -Djava.awt.headless=true  -Dcom.sun.management.jmxremote -Dcom.sun.management.jmxremote.port=8765 -Dcom.sun.management.jmxremote.authenticate=false"
  #Environment="JAVA_OPTS=-Dhttps.protocols=TLSv1,TLSv1.1,TLSv1.2 -Xmx1536m -XX:+DisableExplicitGC -Djava.awt.headless=true"
  TOMCAT_JAVA_OPTS=$(systemctl cat callrec-tomcat | grep Environment=\"JAVA_OPTS= | grep -v \# | tail -n1)

  if [ $(echo ${TOMCAT_JAVA_OPTS} | grep "${JMX_PORT}" | wc -l) -lt 1 ]; then
    echo "     INFO     WEB parameters do not have JMX port set:"
    echo "              ${TOMCAT_JAVA_OPTS}"
  else
    TOMCAT_JAVA_OPTS_OVER=$(grep "^Environment=\"JAVA_OPTS=" ${CALLREC_TOMCAT_SERVICE} | tail -n1)
    TOMCAT_JAVA_OPTS_LINE=$(grep -n "Environment=\"JAVA_OPTS=" ${CALLREC_TOMCAT_SERVICE} | tail -n1 | sed -e 's|:.*$||')
    sed -i "s|^${TOMCAT_JAVA_OPTS_OVER}|#${TOMCAT_JAVA_OPTS_OVER}|" ${CALLREC_TOMCAT_SERVICE}

    for java_param in $(echo ${PARAMS_FOR_JMX}); do
      TOMCAT_JAVA_OPTS_OVER=$(echo ${TOMCAT_JAVA_OPTS_OVER} | sed -e "s|\ ${java_param}||")
    done
    sed -i "${TOMCAT_JAVA_OPTS_LINE}a\\${TOMCAT_JAVA_OPTS_OVER}" ${CALLREC_TOMCAT_SERVICE}
    echo "     INFO     WEB parameters were set to:"
    echo "              ${TOMCAT_JAVA_OPTS_OVER}"
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
    read -p "              Do you wnat to restart now ? [Y/N]   : " restart_web_var
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

function is_finished() {
  echo
  echo "-------------------------------"
  echo "INFO Removal is finished."
  echo "     Database ${DB_HOST} was updated"
  echo "     You can check the service status by command \"qm-services\"."
}

function additional_info() {
  if [ ${WEB_CONFIG_APPLIED} -eq 0 ]; then
    echo
    echo "WEB configuration was not applied. Please finalize the installation manually. "
    echo "          on the web server: Open the FW port: ${JMX_PORT}"
    echo "          on the web server edit the callre-tomcat service. Extend JAVA_OPTS parameters with \" ${PARAMS_FOR_JMX} \"."
  fi

  if [ ${WEB_RESTART_REQUIRED} -eq 0 ]; then
    echo
    echo "WEB application was not restarted. To finish installation restart the web application manually by command: \" systemctl restart callrec-tomcat \""
  fi
}

start_remove
read_install_info
read_config
disable_service
cleanup_db
clean_install_dir
remove_web_jmx_config
restart_web_ui
is_finished
additional_info
