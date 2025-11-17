export namespace main {
	
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

