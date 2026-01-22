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

export namespace config {
	
	export class TrialStatus {
	    isActivated: boolean;
	    isExpired: boolean;
	    remaining: number;
	    days: number;
	    hours: number;
	
	    static createFrom(source: any = {}) {
	        return new TrialStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isActivated = source["isActivated"];
	        this.isExpired = source["isExpired"];
	        this.remaining = source["remaining"];
	        this.days = source["days"];
	        this.hours = source["hours"];
	    }
	}

}

