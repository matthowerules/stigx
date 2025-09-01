export namespace main {
	
	export class ConversionResult {
	    success: boolean;
	    outputPath?: string;
	    errorMessage?: string;
	    inputSize: number;
	    outputSize: number;
	
	    static createFrom(source: any = {}) {
	        return new ConversionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.outputPath = source["outputPath"];
	        this.errorMessage = source["errorMessage"];
	        this.inputSize = source["inputSize"];
	        this.outputSize = source["outputSize"];
	    }
	}
	export class FileInfo {
	    name: string;
	    path: string;
	    size: number;
	    format: string;
	    isValid: boolean;
	    validationError?: string;
	    statistics?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	        this.format = source["format"];
	        this.isValid = source["isValid"];
	        this.validationError = source["validationError"];
	        this.statistics = source["statistics"];
	    }
	}

}

export namespace models {
	
	export class ChecklistItem {
	    id: number;
	    checklistId: number;
	    stigRuleId: number;
	    vulnId: string;
	    status: string;
	    findingDetails?: string;
	    comments?: string;
	    severityOverride?: string;
	    severityJustification?: string;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ChecklistItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.checklistId = source["checklistId"];
	        this.stigRuleId = source["stigRuleId"];
	        this.vulnId = source["vulnId"];
	        this.status = source["status"];
	        this.findingDetails = source["findingDetails"];
	        this.comments = source["comments"];
	        this.severityOverride = source["severityOverride"];
	        this.severityJustification = source["severityJustification"];
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class ImportedSTIG {
	    id: number;
	    name: string;
	    version: string;
	    releaseDate?: string;
	    benchmarkId: string;
	    title?: string;
	    description?: string;
	    filename: string;
	    filePath?: string;
	    fileHash?: string;
	    ruleCount: number;
	    // Go type: time
	    importedAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ImportedSTIG(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.version = source["version"];
	        this.releaseDate = source["releaseDate"];
	        this.benchmarkId = source["benchmarkId"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.filename = source["filename"];
	        this.filePath = source["filePath"];
	        this.fileHash = source["fileHash"];
	        this.ruleCount = source["ruleCount"];
	        this.importedAt = this.convertValues(source["importedAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class ImportedSTIGRule {
	    id: number;
	    stigId: number;
	    ruleId: string;
	    versionId?: string;
	    title?: string;
	    description?: string;
	    severity?: string;
	    weight?: number;
	    groupId?: string;
	    groupTitle?: string;
	    checkContent?: string;
	    fixText?: string;
	    cciRefs?: string;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ImportedSTIGRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.stigId = source["stigId"];
	        this.ruleId = source["ruleId"];
	        this.versionId = source["versionId"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.severity = source["severity"];
	        this.weight = source["weight"];
	        this.groupId = source["groupId"];
	        this.groupTitle = source["groupTitle"];
	        this.checkContent = source["checkContent"];
	        this.fixText = source["fixText"];
	        this.cciRefs = source["cciRefs"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
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
	export class ImportedSTIGSummary {
	    id: number;
	    name: string;
	    version: string;
	    releaseDate?: string;
	    title?: string;
	    ruleCount: number;
	    // Go type: time
	    importedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ImportedSTIGSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.version = source["version"];
	        this.releaseDate = source["releaseDate"];
	        this.title = source["title"];
	        this.ruleCount = source["ruleCount"];
	        this.importedAt = this.convertValues(source["importedAt"], null);
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
	export class TargetData {
	    target_type?: string;
	    host_name?: string;
	    ip_address?: string;
	    mac_address?: string;
	    fqdn?: string;
	    comments?: string;
	    role?: string;
	    is_web_database?: boolean;
	    technology_area?: string;
	    web_db_site?: string;
	    web_db_instance?: string;
	
	    static createFrom(source: any = {}) {
	        return new TargetData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.target_type = source["target_type"];
	        this.host_name = source["host_name"];
	        this.ip_address = source["ip_address"];
	        this.mac_address = source["mac_address"];
	        this.fqdn = source["fqdn"];
	        this.comments = source["comments"];
	        this.role = source["role"];
	        this.is_web_database = source["is_web_database"];
	        this.technology_area = source["technology_area"];
	        this.web_db_site = source["web_db_site"];
	        this.web_db_instance = source["web_db_instance"];
	    }
	}
	export class UserChecklist {
	    id: number;
	    name: string;
	    description?: string;
	    targetData?: string;
	    createdFromStigs: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new UserChecklist(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.targetData = source["targetData"];
	        this.createdFromStigs = source["createdFromStigs"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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

export namespace services {
	
	export class ChecklistCreationResult {
	    checklistId: number;
	    name: string;
	    format: string;
	    filePath: string;
	    ruleCount: number;
	    stigCount: number;
	
	    static createFrom(source: any = {}) {
	        return new ChecklistCreationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.checklistId = source["checklistId"];
	        this.name = source["name"];
	        this.format = source["format"];
	        this.filePath = source["filePath"];
	        this.ruleCount = source["ruleCount"];
	        this.stigCount = source["stigCount"];
	    }
	}
	export class ChecklistItemWithRule {
	    Item?: models.ChecklistItem;
	    Rule?: models.ImportedSTIGRule;
	    STIG?: models.ImportedSTIG;
	
	    static createFrom(source: any = {}) {
	        return new ChecklistItemWithRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Item = this.convertValues(source["Item"], models.ChecklistItem);
	        this.Rule = this.convertValues(source["Rule"], models.ImportedSTIGRule);
	        this.STIG = this.convertValues(source["STIG"], models.ImportedSTIG);
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
	export class ChecklistWithItems {
	    Checklist?: models.UserChecklist;
	    Items: ChecklistItemWithRule[];
	    STIGs: models.ImportedSTIG[];
	
	    static createFrom(source: any = {}) {
	        return new ChecklistWithItems(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Checklist = this.convertValues(source["Checklist"], models.UserChecklist);
	        this.Items = this.convertValues(source["Items"], ChecklistItemWithRule);
	        this.STIGs = this.convertValues(source["STIGs"], models.ImportedSTIG);
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
	export class CreateChecklistRequest {
	    name: string;
	    description?: string;
	    stigIds: number[];
	    targetData?: models.TargetData;
	    format: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateChecklistRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.stigIds = source["stigIds"];
	        this.targetData = this.convertValues(source["targetData"], models.TargetData);
	        this.format = source["format"];
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

