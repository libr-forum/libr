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

