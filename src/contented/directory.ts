import * as _ from 'lodash';
import {ApiDef} from './api_def';

export class Directory {
    public contents: Array<string>;
    public total: number;
    public path: string;
    public id: string;

    constructor(dir: any) {
        this.contents = _.get(dir, 'contents') || [];
        this.total = _.get(dir, 'total') || 0;
        this.path = _.get(dir, 'path') || '';
        this.id = _.get(dir, 'id') || '';
    }

    public getContentList() {
        return _.map(this.contents, c => {
            return ApiDef.base + this.trail(this.path, '/') + (c || '');
        });
    }

    public trail(path: string, whatWith: string) {
        if (path[path.length - 1] !== whatWith) {
            return path + whatWith;
        }
        return path;
    }
}
