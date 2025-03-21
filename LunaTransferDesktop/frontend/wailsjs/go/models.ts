export namespace main {
	
	export class FileItem {
	    name: string;
	    path: string;
	    isDirectory: boolean;
	    size: number;
	    modified: string;
	
	    static createFrom(source: any = {}) {
	        return new FileItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.isDirectory = source["isDirectory"];
	        this.size = source["size"];
	        this.modified = source["modified"];
	    }
	}

}

