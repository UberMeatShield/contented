import * as _ from 'lodash';
import {Media} from './media';
import {ApiDef} from './api_def';


function trail(path: string, whatWith: string) {
    if (path[path.length - 1] !== whatWith) {
        return path + whatWith;
    }
    return path;
}

export enum LoadStates {
    NotLoaded,
    Loading,
    Partial,
    Complete
}

export class Container {
    public contents: Array<Media>;
    public total: number;
    public count: number;
    public path: string;
    public name: string;
    public id: string;
    public previewUrl: string;

    // Set on the initial content loads
    public loadState: LoadStates = LoadStates.NotLoaded;

    // All potential items that can be rendered from the contents
    public renderable: Array<Media>;
    public visibleSet: Array<Media> = [];

    constructor(dir: any) {
        this.total = _.get(dir, 'total') || 0;
        this.id = _.get(dir, 'id') || '';
        this.name = _.get(dir, 'name') || '';
        this.previewUrl = _.get(dir, 'preview_src') || '';
        this.setContents(this.buildImgs(_.get(dir, 'contents') || []));
    }

    // For use in determining what should actually be visible at any time
    public getIntervalAround(currentItem: Media, requestedVisible: number = 4, before: number = 0) {
        this.visibleSet = null;

        let items = this.getContentList() || [];
        let start = 0;
        let max = requestedVisible < items.length ? requestedVisible : items.length;

        // Need to look it up by ID
        if (currentItem) {
            start = this.indexOf(currentItem);
            start = start >= 0 ? start : 0;
            start = (before && (start - before > 0)) ? (start - before) : 0;
            // console.log("What is the start for the loading interval", currentItem.id, start, max);
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

    public indexOf(item: Media, contents: Array<Media> = null) {
        contents = contents || this.getContentList() || [];
        if ( item && contents ) {
            return _.findIndex(contents, {id: item.id});
        }
        return -1;
    }

    public buildImgs(imgData: Array<any>) {
        return _.map(imgData, data => new Media(data));
    }

    public setContents(contents: Array<Media>) {
        this.contents = _.sortBy(_.uniqBy(contents || [], 'id'), 'idx');
        this.count = this.contents.length;
        this.renderable = null;

        if (this.count === this.total) {
            this.loadState = LoadStates.Complete;
        } else if (this.loadState === LoadStates.Loading) {
            this.loadState = LoadStates.Partial;
        }
    }

    public addContents(contents: Array<Media>) {
        let sorted = _.sortBy((this.contents || []).concat(contents), 'idx');
        console.log("What is going on", sorted);
        this.setContents(sorted);
    }

    // This is the actual URL you can get a pointer to for the scroll / load
    public getContentList() {
        if (!this.renderable) {
            this.renderable = _.map(this.contents, (c: Media) => {
                return c;
            });
        }
        return this.renderable || [];
    }
}
