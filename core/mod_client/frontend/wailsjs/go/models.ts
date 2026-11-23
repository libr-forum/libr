export namespace app {
	
	export class Community {
	    name: string;
	    is_mod: boolean;
	    mods: number;
	    threshold: number;
	    mod_keys: string[];
	
	    static createFrom(source: any = {}) {
	        return new Community(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.is_mod = source["is_mod"];
	        this.mods = source["mods"];
	        this.threshold = source["threshold"];
	        this.mod_keys = source["mod_keys"];
	    }
	}
	export class Outcome {
	    status: string;
	    hash?: string;
	    ts?: number;
	    id?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new Outcome(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.hash = source["hash"];
	        this.ts = source["ts"];
	        this.id = source["id"];
	        this.error = source["error"];
	    }
	}
	export class PendingInfo {
	    id: string;
	    key: string;
	    kind: string;
	    community: string;
	    content: string;
	    reason?: string;
	    created: number;
	    approved: number;
	    rejected: number;
	    waiting: number;
	    status: string;
	    hash?: string;
	
	    static createFrom(source: any = {}) {
	        return new PendingInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.key = source["key"];
	        this.kind = source["kind"];
	        this.community = source["community"];
	        this.content = source["content"];
	        this.reason = source["reason"];
	        this.created = source["created"];
	        this.approved = source["approved"];
	        this.rejected = source["rejected"];
	        this.waiting = source["waiting"];
	        this.status = source["status"];
	        this.hash = source["hash"];
	    }
	}
	export class Post {
	    hash: string;
	    msg_id: string;
	    community: string;
	    author: string;
	    ts: number;
	    content: string;
	    votes: number;
	    mine: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Post(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hash = source["hash"];
	        this.msg_id = source["msg_id"];
	        this.community = source["community"];
	        this.author = source["author"];
	        this.ts = source["ts"];
	        this.content = source["content"];
	        this.votes = source["votes"];
	        this.mine = source["mine"];
	    }
	}
	export class Status {
	    online: boolean;
	    peers: number;
	    routing_table: number;
	    reservations: number;
	
	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.online = source["online"];
	        this.peers = source["peers"];
	        this.routing_table = source["routing_table"];
	        this.reservations = source["reservations"];
	    }
	}

}

export namespace main {
	
	export class QueueItem {
	    key: string;
	    kind: string;
	    community: string;
	    content: string;
	    reason: string;
	    received: number;
	    decided: boolean;
	    approved: boolean;
	
	    static createFrom(source: any = {}) {
	        return new QueueItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.kind = source["kind"];
	        this.community = source["community"];
	        this.content = source["content"];
	        this.reason = source["reason"];
	        this.received = source["received"];
	        this.decided = source["decided"];
	        this.approved = source["approved"];
	    }
	}
	export class StorageStatus {
	    enabled: boolean;
	    running: boolean;
	    peer_id: string;
	
	    static createFrom(source: any = {}) {
	        return new StorageStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.running = source["running"];
	        this.peer_id = source["peer_id"];
	    }
	}

}

export namespace models {
	
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

}

