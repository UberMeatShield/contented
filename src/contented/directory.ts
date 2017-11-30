import * as _ from 'lodash';
import {ApiDef} from './api_def';

export class Directory {
    public contents: Array<string>;
    public total: number;
    public count: number;
    public path: string;
    public id: string;

    // All potential items that can be rendered from the contents
    public renderable: Array<string>;
    public visibleSet: Array<string> = [];

    constructor(dir: any) {
        this.total = _.get(dir, 'total') || 0;
        this.path = _.get(dir, 'path') || '';
        this.id = _.get(dir, 'id') || '';
        this.setContents(_.get(dir, 'contents'));
    }

    // For use in determining what should actually be visible at any time
    public getIntervalAround(currentItem: string = '', max: number = 4, before: number = 0) {
         this.visibleSet = null;

         let items = this.getContentList() || [];
         let idx = 0;
         if (currentItem) {
             idx = _.indexOf(items, currentItem);
             idx = idx >= 0 ? idx : 0;
             idx = before && (idx - before > 0) ? idx - before : 0;
         }
         let end = idx + (max >= 1 ? max : 4);
         this.visibleSet = items.slice(idx, end < items.length ? end : items.length) || [];
         return this.visibleSet;
    }

    public setContents(contents: Array<string>) {
        this.contents = _.isArray(contents) ? contents : [];
        this.count = this.contents.length;
        this.renderable = null;
    }

    // This is the actual URL you can get a pointer to for the scroll / load
    public getContentList() {
        if (!this.renderable) {
            this.renderable = _.map(this.contents, c => {
                return ApiDef.base + this.trail(this.path, '/') + (c || '');
            });
        }
        return this.renderable || [];
    }

    public trail(path: string, whatWith: string) {
        if (path[path.length - 1] !== whatWith) {
            return path + whatWith;
        }
        return path;
    }
}
