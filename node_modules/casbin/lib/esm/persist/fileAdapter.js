import { Helper } from './helper';
import { arrayToString, readFile, writeFile } from '../util';
/**
 * FileAdapter is the file adapter for Casbin.
 * It can load policy from file or save policy to file.
 */
export class FileAdapter {
    /**
     * FileAdapter is the constructor for FileAdapter.
     * @param {string} filePath filePath the path of the policy file.
     */
    constructor(filePath) {
        this.filePath = filePath;
    }
    async loadPolicy(model) {
        if (!this.filePath) {
            // throw new Error('invalid file path, file path cannot be empty');
            return;
        }
        await this.loadPolicyFile(model, Helper.loadPolicyLine);
    }
    async loadPolicyFile(model, handler) {
        const bodyBuf = await readFile(this.filePath);
        const lines = bodyBuf.toString().split('\n');
        lines.forEach((n, index) => {
            if (!n) {
                return;
            }
            handler(n, model);
        });
    }
    /**
     * savePolicy saves all policy rules to the storage.
     */
    async savePolicy(model) {
        if (!this.filePath) {
            // throw new Error('invalid file path, file path cannot be empty');
            return false;
        }
        let result = '';
        const pList = model.model.get('p');
        if (!pList) {
            return false;
        }
        pList.forEach((n) => {
            n.policy.forEach((m) => {
                result += n.key + ', ';
                result += arrayToString(m);
                result += '\n';
            });
        });
        const gList = model.model.get('g');
        if (!gList) {
            return false;
        }
        gList.forEach((n) => {
            n.policy.forEach((m) => {
                result += n.key + ', ';
                result += arrayToString(m);
                result += '\n';
            });
        });
        await this.savePolicyFile(result.trim());
        return true;
    }
    async savePolicyFile(text) {
        await writeFile(this.filePath, text);
    }
    /**
     * addPolicy adds a policy rule to the storage.
     */
    async addPolicy(sec, ptype, rule) {
        throw new Error('not implemented');
    }
    /**
     * addPolicies adds policy rules to the storage.
     This is part of the Auto-Save feature.
     */
    async addPolicies(sec, ptype, rules) {
        throw new Error('not implemented');
    }
    /**
     * UpdatePolicy updates a policy rule from storage.
     * This is part of the Auto-Save feature.
     */
    updatePolicy(sec, ptype, oldRule, newRule) {
        throw new Error('not implemented');
    }
    /**
     * removePolicy removes a policy rule from the storage.
     */
    async removePolicy(sec, ptype, rule) {
        throw new Error('not implemented');
    }
    /**
     * removePolicies removes policy rules from the storage.
     * This is part of the Auto-Save feature.
     */
    async removePolicies(sec, ptype, rules) {
        throw new Error('not implemented');
    }
    /**
     * removeFilteredPolicy removes policy rules that match the filter from the storage.
     */
    async removeFilteredPolicy(sec, ptype, fieldIndex, ...fieldValues) {
        throw new Error('not implemented');
    }
}
