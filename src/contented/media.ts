import * as _ from 'lodash';
import {ApiDef} from './api_def';
import {Screen} from './screen';

export class Media {
    public id: string;
    public src: string;
    public idx: number;

    // Awkward that buffalo makes the API use container_id like the DB
    // side of things and in url params by default.  So I guess mixed
    // cases it is.
    public content_type: string;
    public container_id: string;

    public previewUrl: string;
    public fullUrl: string;
    public screens: Array<Screen>;

    constructor(obj: any = {}) {
        this.fromJson(obj);
    }

    public fromJson(raw: any) {
        if (raw) {
            Object.assign(this, raw);
            this.links();
            this.screens = _.map(raw.screens, s => new Screen(s));
        }
    }

    public isImage() {
        return this.content_type ? !!(this.content_type.match("image")) : false;
    }

    public isVideo() {
        return this.content_type ? !!(this.content_type.match("video")) : false;
    }

    public links() {
        this.previewUrl = `${ApiDef.contented.preview}${this.id}`;
        this.fullUrl = `${ApiDef.contented.view}${this.id}`;
    }
}
