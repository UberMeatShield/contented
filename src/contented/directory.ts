import * as _ from 'lodash';
import {ApiDef} from './api_def';


function trail(path: string, whatWith: string) {
    if (path[path.length - 1] !== whatWith) {
        return path + whatWith;
    }
    return path;
}

export class ImgContainer {
    public id: number;
    public src: string;
    public fullPath: string;

    constructor(id: number, src: string, path: string = null) {
        this.id = id;
        this.src = src;

        if (path) {
            this.setPath(path);
        }
    }

    public setPath(path: string) {
        this.fullPath = `${ApiDef.base}${trail(path, '/')}${this.src}`;
    }
}

export class Directory {
    public contents: Array<ImgContainer>;
    public total: number;
    public count: number;
    public path: string;
    public id: string;

    // All potential items that can be rendered from the contents
    public renderable: Array<ImgContainer>;
    public visibleSet: Array<ImgContainer> = [];

    constructor(dir: any) {
        this.total = _.get(dir, 'total') || 0;
        this.path = _.get(dir, 'path') || '';
        this.id = _.get(dir, 'id') || '';

        this.setContents(this.buildImgs(_.get(dir, 'contents') || []));
    }

    // For use in determining what should actually be visible at any time
    public getIntervalAround(currentItem: ImgContainer, requestedVisible: number = 4, before: number = 0) {
        this.visibleSet = null;

        let items = this.getContentList() || [];
        let start = 0;
        let max = requestedVisible < items.length ? requestedVisible : items.length;

        // Need to look it up by ID
        if (currentItem) {
            start = this.indexOf(currentItem);
            start = start >= 0 ? start : 0;
            start = (before && (start - before > 0)) ? (start - before) : 0;

            console.log("What is the start", currentItem.id, start, max);
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

    public indexOf(item: ImgContainer, contents: Array<ImgContainer> = null) {
        contents = contents || this.getContentList() || [];
        if ( item && contents ) {
            return _.findIndex(contents, {id: item.id});
        }
        return -1;
    }

    public buildImgs(imgData: Array<any>) {
        return _.map(imgData, data => {
            return new ImgContainer(data.id, data.src, this.path);
        });
    }

    public setContents(contents: Array<ImgContainer>) {
        this.contents = _.sortBy(_.uniqBy(contents || [], 'id'), 'id');
        this.count = this.contents.length;
        this.renderable = null;
    }

    public addContents(contents: Array<ImgContainer>) {
        this.setContents((this.contents || []).concat(contents));
    }

    // This is the actual URL you can get a pointer to for the scroll / load
    public getContentList() {
        if (!this.renderable) {
            this.renderable = _.map(this.contents, (c: ImgContainer) => {
                return c;
            });
        }
        return this.renderable || [];
    }
}
