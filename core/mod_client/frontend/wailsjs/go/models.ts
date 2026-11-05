export namespace main {
	
	export class PendingItemStat {
	    msg_sign: string;
	    ts: number;
	    content: string;
	    reason: string;
	    approved: number;
	    rejected: number;
	    awaiting: number;
	    is_image: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PendingItemStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.msg_sign = source["msg_sign"];
	        this.ts = source["ts"];
	        this.content = source["content"];
	        this.reason = source["reason"];
	        this.approved = source["approved"];
	        this.rejected = source["rejected"];
	        this.awaiting = source["awaiting"];
	        this.is_image = source["is_image"];
	    }
	}
	export class PendingModerationStats {
	    items: PendingItemStat[];
	    cron_active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PendingModerationStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], PendingItemStat);
	        this.cron_active = source["cron_active"];
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

export namespace models {
	
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
	export class ModConfig {
	    forbidden: string[];
	    thresholds: string;
	    image_auto_mod?: boolean;
	    llm_enabled: boolean;
	    llm_provider: string;
	    llm_model: string;
	
	    static createFrom(source: any = {}) {
	        return new ModConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.forbidden = source["forbidden"];
	        this.thresholds = source["thresholds"];
	        this.image_auto_mod = source["image_auto_mod"];
	        this.llm_enabled = source["llm_enabled"];
	        this.llm_provider = source["llm_provider"];
	        this.llm_model = source["llm_model"];
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
	    type?: string;
	
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
	        this.type = source["type"];
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
	    type?: string;
	
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
	        this.type = source["type"];
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
	export class SendResult {
	    status: string;
	    mod_certs: ModCert[];
	    sign: string;
	    ts: number;
	
	    static createFrom(source: any = {}) {
	        return new SendResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.mod_certs = this.convertValues(source["mod_certs"], ModCert);
	        this.sign = source["sign"];
	        this.ts = source["ts"];
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

