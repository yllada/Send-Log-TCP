// ================================================================================
// SYSLOG CONSTANTS
// Single Responsibility: Defines RFC 5424 facility and severity options
// ================================================================================

export interface SelectOption {
  value: number;
  label: string;
}

// Facility codes according to RFC 5424
export const FACILITY_OPTIONS: SelectOption[] = [
  { value: 0, label: "0 - kernel" },
  { value: 1, label: "1 - user" },
  { value: 2, label: "2 - mail" },
  { value: 3, label: "3 - system daemon" },
  { value: 4, label: "4 - security/auth" },
  { value: 5, label: "5 - syslogd" },
  { value: 6, label: "6 - line printer" },
  { value: 7, label: "7 - network news" },
  { value: 8, label: "8 - UUCP" },
  { value: 9, label: "9 - clock daemon" },
  { value: 10, label: "10 - security/auth" },
  { value: 11, label: "11 - FTP daemon" },
  { value: 12, label: "12 - NTP" },
  { value: 13, label: "13 - log audit" },
  { value: 14, label: "14 - log alert" },
  { value: 15, label: "15 - clock daemon" },
  { value: 16, label: "16 - local0" },
  { value: 17, label: "17 - local1" },
  { value: 18, label: "18 - local2" },
  { value: 19, label: "19 - local3" },
  { value: 20, label: "20 - local4" },
  { value: 21, label: "21 - local5" },
  { value: 22, label: "22 - local6" },
  { value: 23, label: "23 - local7" },
];

// Severity levels according to RFC 5424
export const SEVERITY_OPTIONS: SelectOption[] = [
  { value: 0, label: "0 - Emergency" },
  { value: 1, label: "1 - Alert" },
  { value: 2, label: "2 - Critical" },
  { value: 3, label: "3 - Error" },
  { value: 4, label: "4 - Warning" },
  { value: 5, label: "5 - Notice" },
  { value: 6, label: "6 - Informational" },
  { value: 7, label: "7 - Debug" },
];

// Default ports
export const DEFAULT_SYSLOG_PORT = "514";
export const DEFAULT_SYSLOG_TLS_PORT = "6514";
