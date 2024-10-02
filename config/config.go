package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func GetConfig() (string, string) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	address := os.Getenv("LOG_AGENT_ADDRESS")
	if address == "" {
		address = "127.0.0.1:7003"
	}

	logMessage := os.Getenv("LOG_MESSAGE")
	if logMessage == "" {
		logMessage = `679 <12>1 2022-04-11T10:41:05.300Z esmc ERAServer 11804 - - {"event_type":"FirewallAggregated_Event","ipv4":"192.168.1.7","hostname":"archivos","source_uuid":"c252c742-22e4-4b4c-af89-6216dc200276","occured":"11-Apr-2022 10:45:38","severity":"Warning","event":"Security vulnerability exploitation attempt","source_address":"95.216.242.119","source_address_type":"IPv4","source_port":53344,"target_address":"192.168.1.7","target_address_type":"IPv4","target_port":3389,"protocol":"TCP","action":"Blocked","handled":true,"account":"NT AUTHORITY\\Servicio de red","process_name":"C:\\Windows\\System32\\svchost.exe","inbound":true,"threat_name":"EsetIpBlacklist","aggregate_count":1}`
	}

	return address, logMessage
}
