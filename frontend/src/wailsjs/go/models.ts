export namespace main {
	
	export class ContinuousSendConfig {
	    Address: string;
	    Port: string;
	    Protocol: string;
	    Message: string;
	    FramingMethod: string;
	    Facility: number;
	    Severity: number;
	    Hostname: string;
	    Appname: string;
	    UseRFC5424: boolean;
	    UseTLS: boolean;
	    TLSVerify: boolean;
	    CACertPath: string;
	    ClientCertPath: string;
	    ClientKeyPath: string;
	    Duration: number;
	    MessagesPerSec: number;
	    MaxMessages: number;
	    RandomizeData: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ContinuousSendConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Address = source["Address"];
	        this.Port = source["Port"];
	        this.Protocol = source["Protocol"];
	        this.Message = source["Message"];
	        this.FramingMethod = source["FramingMethod"];
	        this.Facility = source["Facility"];
	        this.Severity = source["Severity"];
	        this.Hostname = source["Hostname"];
	        this.Appname = source["Appname"];
	        this.UseRFC5424 = source["UseRFC5424"];
	        this.UseTLS = source["UseTLS"];
	        this.TLSVerify = source["TLSVerify"];
	        this.CACertPath = source["CACertPath"];
	        this.ClientCertPath = source["ClientCertPath"];
	        this.ClientKeyPath = source["ClientKeyPath"];
	        this.Duration = source["Duration"];
	        this.MessagesPerSec = source["MessagesPerSec"];
	        this.MaxMessages = source["MaxMessages"];
	        this.RandomizeData = source["RandomizeData"];
	    }
	}
	export class ContinuousStats {
	    totalSent: number;
	    totalErrors: number;
	    currentRate: number;
	    elapsedSeconds: number;
	    isRunning: boolean;
	    startTime: number;
	    targetRate: number;
	    duration: number;
	
	    static createFrom(source: any = {}) {
	        return new ContinuousStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalSent = source["totalSent"];
	        this.totalErrors = source["totalErrors"];
	        this.currentRate = source["currentRate"];
	        this.elapsedSeconds = source["elapsedSeconds"];
	        this.isRunning = source["isRunning"];
	        this.startTime = source["startTime"];
	        this.targetRate = source["targetRate"];
	        this.duration = source["duration"];
	    }
	}
	export class SyslogConfig {
	    Address: string;
	    Port: string;
	    Protocol: string;
	    Messages: string[];
	    FramingMethod: string;
	    Facility: number;
	    Severity: number;
	    Hostname: string;
	    Appname: string;
	    UseRFC5424: boolean;
	    UseTLS: boolean;
	    TLSVerify: boolean;
	    CACertPath: string;
	    ClientCertPath: string;
	    ClientKeyPath: string;
	
	    static createFrom(source: any = {}) {
	        return new SyslogConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Address = source["Address"];
	        this.Port = source["Port"];
	        this.Protocol = source["Protocol"];
	        this.Messages = source["Messages"];
	        this.FramingMethod = source["FramingMethod"];
	        this.Facility = source["Facility"];
	        this.Severity = source["Severity"];
	        this.Hostname = source["Hostname"];
	        this.Appname = source["Appname"];
	        this.UseRFC5424 = source["UseRFC5424"];
	        this.UseTLS = source["UseTLS"];
	        this.TLSVerify = source["TLSVerify"];
	        this.CACertPath = source["CACertPath"];
	        this.ClientCertPath = source["ClientCertPath"];
	        this.ClientKeyPath = source["ClientKeyPath"];
	    }
	}
	export class SyslogResponse {
	    sentMessages: string[];
	    errors: string[];
	
	    static createFrom(source: any = {}) {
	        return new SyslogResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sentMessages = source["sentMessages"];
	        this.errors = source["errors"];
	    }
	}

}

