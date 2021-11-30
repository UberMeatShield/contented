import * as _ from 'lodash';
import {ApiDef} from './api_def';

export class Media {
    public id: string;
    public src: string;
    public idx: number;
    public content_type: string;
    public container_id: string;

    public previewUrl: string;
    public fullUrl: string;

    constructor(obj: any = {}) {
        this.fromJson(obj);
    }

    public fromJson(raw: any) {
        if (raw) {
            Object.assign(this, raw);
            this.links();
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
