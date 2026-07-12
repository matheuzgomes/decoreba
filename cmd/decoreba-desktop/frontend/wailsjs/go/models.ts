export namespace core {
	
	export class Command {
	    id: string;
	    context: string;
	    title: string;
	    command: string;
	    tags?: string[];
	    notes?: string;
	    usage_count: number;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Command(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.context = source["context"];
	        this.title = source["title"];
	        this.command = source["command"];
	        this.tags = source["tags"];
	        this.notes = source["notes"];
	        this.usage_count = source["usage_count"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
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
	export class Settings {
	    width: number;
	    height: number;
	    font_scale: number;
	    always_on_top: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.width = source["width"];
	        this.height = source["height"];
	        this.font_scale = source["font_scale"];
	        this.always_on_top = source["always_on_top"];
	    }
	}

}

export namespace main {
	
	export class searchResult {
	    cmd: core.Command;
	    score: number;
	
	    static createFrom(source: any = {}) {
	        return new searchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cmd = this.convertValues(source["cmd"], core.Command);
	        this.score = source["score"];
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

