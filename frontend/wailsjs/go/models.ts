export namespace core {
	
	export class Metadata {
	    ID: string;
	    Name: string;
	    Extensions: string[];
	    FramesPerSecond: number;
	    ScreenWidth: number;
	    ScreenHeight: number;
	
	    static createFrom(source: any = {}) {
	        return new Metadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Extensions = source["Extensions"];
	        this.FramesPerSecond = source["FramesPerSecond"];
	        this.ScreenWidth = source["ScreenWidth"];
	        this.ScreenHeight = source["ScreenHeight"];
	    }
	}

}

export namespace game {
	
	export class Game {
	    id: string;
	    title: string;
	    system: string;
	    path: string;
	    artworkPath?: string;
	    // Go type: time
	    addedAt: any;
	    // Go type: time
	    lastPlayedAt?: any;
	    playTimeSeconds: number;
	
	    static createFrom(source: any = {}) {
	        return new Game(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.system = source["system"];
	        this.path = source["path"];
	        this.artworkPath = source["artworkPath"];
	        this.addedAt = this.convertValues(source["addedAt"], null);
	        this.lastPlayedAt = this.convertValues(source["lastPlayedAt"], null);
	        this.playTimeSeconds = source["playTimeSeconds"];
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

