export namespace app {
	
	export class ExtractResult {
	    success: boolean;
	    recordCount: number;
	    outputPath: string;
	    errorMessage?: string;
	    records?: extractor.Record[];
	
	    static createFrom(source: any = {}) {
	        return new ExtractResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.recordCount = source["recordCount"];
	        this.outputPath = source["outputPath"];
	        this.errorMessage = source["errorMessage"];
	        this.records = this.convertValues(source["records"], extractor.Record);
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

export namespace extractor {
	
	export class Record {
	    defendant: string;
	    idNumber: string;
	    request: string;
	    factsReason: string;
	
	    static createFrom(source: any = {}) {
	        return new Record(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.defendant = source["defendant"];
	        this.idNumber = source["idNumber"];
	        this.request = source["request"];
	        this.factsReason = source["factsReason"];
	    }
	}

}

