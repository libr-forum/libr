export namespace models {
	
	export class ModConfig {
	    forbidden: string[];
	    thresholds: string;
	
	    static createFrom(source: any = {}) {
	        return new ModConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.forbidden = source["forbidden"];
	        this.thresholds = source["thresholds"];
	    }
	}
	export class ModLogEntry {
	    public_key: string;
	    content: string;
	    timestamp: number;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new ModLogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.public_key = source["public_key"];
	        this.content = source["content"];
	        this.timestamp = source["timestamp"];
	        this.status = source["status"];
	    }
	}

}

export namespace types {
	
	export class ModCert {
	    sign: string;
	    public_key: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new ModCert(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sign = source["sign"];
	        this.public_key = source["public_key"];
	        this.status = source["status"];
	    }
	}
	export class Msg {
	    content: string;
	    ts: number;
	
	    static createFrom(source: any = {}) {
	        return new Msg(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.content = source["content"];
	        this.ts = source["ts"];
	    }
	}
	export class MsgCert {
	    public_key: string;
	    msg: Msg;
	    mod_certs: ModCert[];
	    sign: string;
	    reason?: string;
	
	    static createFrom(source: any = {}) {
	        return new MsgCert(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.public_key = source["public_key"];
	        this.msg = this.convertValues(source["msg"], Msg);
	        this.mod_certs = this.convertValues(source["mod_certs"], ModCert);
	        this.sign = source["sign"];
	        this.reason = source["reason"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RetMsgCert {
	    public_key: string;
	    msg: Msg;
	    mod_certs: ModCert[];
	    sign: string;
	    deleted: string;
	
	    static createFrom(source: any = {}) {
	        return new RetMsgCert(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.public_key = source["public_key"];
	        this.msg = this.convertValues(source["msg"], Msg);
	        this.mod_certs = this.convertValues(source["mod_certs"], ModCert);
	        this.sign = source["sign"];
	        this.deleted = source["deleted"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

