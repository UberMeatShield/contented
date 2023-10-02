import * as _ from 'lodash';
import {ApiDef} from './api_def';
import {Screen} from './screen';

// Why does a TAG have an id?!?!  Because goBuffalo really likes the id field.
export class Tag {
    public id: string;

    constructor(tagName: string) {
        this.id = tagName;
    }
}

export class Content {
    public id: string;
    public src: string;
    public preview: string; // Name of the preview, if not set we do not have one.
    public idx: number;
    public description: string = "";

    // Awkward that buffalo makes the API use container_id like the DB
    // side of things and in url params by default.  So I guess mixed
    // cases it is.
    public content_type: string;
    public container_id: string;
    public size: number;

    public previewUrl: string;
    public fullUrl: string;
    public screens: Array<Screen>;
    public tags: Array<Tag>;
    public meta: string;

    public fullText: string|undefined = undefined;

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

    public isText() {
        return this.content_type ? !!(this.content_type.match("text")) : false;
    }

    getVideoInfo() {
        if (this.isVideo() && this.meta) {
            let videoInfo = JSON.parse(this.meta);
            console.log("Video info", videoInfo)
        } 
        return undefined
    }

    // Images will just work as a preview source, but video (with no preview) and 
    // text and zips etc should use a content_type based style preview.  This prevents
    // broken image links when no previews are found for these types.
    public shouldUseTypedPreview() {
        if (_.isEmpty(this.preview)) {
            if (this.isImage()) {
                return "";          // image we can just display the image (maybe large)
            } else if (this.isVideo()) {
                return "videocam";  // material icon
            } else if (this.isText()) {
                return "article";  // material icon
            }
        }
        return "";
    }

    public links() {
        this.previewUrl = `${ApiDef.contented.preview}${this.id}`;
        this.fullUrl = `${ApiDef.contented.view}${this.id}`;
    }
}
