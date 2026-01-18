export namespace app {
	
	export class ExtractResult {
	    success: boolean;
	    recordCount: number;
	    outputPath: string;
	    errorMessage?: string;
	    records?: any[];
	    fieldLabels?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new ExtractResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.recordCount = source["recordCount"];
	        this.outputPath = source["outputPath"];
	        this.errorMessage = source["errorMessage"];
	        this.records = source["records"];
	        this.fieldLabels = source["fieldLabels"];
	    }
	}
	export class FieldOption {
	    key: string;
	    label: string;
	
	    static createFrom(source: any = {}) {
	        return new FieldOption(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.label = source["label"];
	    }
	}

}

