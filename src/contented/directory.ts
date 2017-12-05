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
    public getIntervalAround(currentItem: string = '', requestedVisible: number = 4, before: number = 0) {
        this.visibleSet = null;

        let items = this.getContentList() || [];
        let start = 0;
        let max = requestedVisible < items.length ? requestedVisible : items.length;

        if (currentItem) {
            start = _.indexOf(items, currentItem);
            start = start >= 0 ? start : 0;
            start = before && (start - before > 0) ? start - before : 0;
        }
        let end = start + (max >= 1 ? max : 4);
            end = end < items.length ? end : items.length;
        let interval = (end - start);
        if (interval < max) {
            start = start - (max - interval);
        }
        this.visibleSet = items.slice(start, end) || [];
        return this.visibleSet;
    }

    public indexOf(item: string) {
        return _.indexOf(this.getContentList(), item);
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
