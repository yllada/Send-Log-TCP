export namespace main {
	
	export class SyslogConfig {
	    Address: string;
	    Port: string;
	    Protocol: string;
	    Messages: string[];
	
	    static createFrom(source: any = {}) {
	        return new SyslogConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Address = source["Address"];
	        this.Port = source["Port"];
	        this.Protocol = source["Protocol"];
	        this.Messages = source["Messages"];
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

