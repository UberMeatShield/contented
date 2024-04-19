import * as _ from 'lodash';
import {ApiDef} from './api_def';
import {Screen} from './screen';

// Why does a TAG have an id?!?!  Because goBuffalo really likes the id field.
export class Tag {
    public id: string;
    public tag_type: string;

    constructor(obj: any) {
        if (typeof obj == 'string') {
            this.id = obj;
        } else {
            Object.assign(this, obj);
        }
    }

    isProblem() {
        let arr = this.id ? this.id.split(" ") : [this.id];
        if (arr.length > 1) {
           return true; 
        }
        return false;
    }
}

export interface VSCodeChange {
    value: string;
    tags: Array<string>;
    force?: boolean;
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
    public size: number;  // size in bytes

    // Only defined currently on video
    public encoding: string | undefined;

    public previewUrl: string;
    public fullUrl: string;
    public screens: Array<Screen>;
    public tags: Array<Tag>;
    public meta: string;

    public fullText: string|undefined = undefined;
    public videoInfo?: VideoCodecInfo;

    public created_at: Date;
    public updated_at: Date;

    constructor(obj: any = {}) {
        this.fromJson(obj);
    }

    public fromJson(raw: any) {
        if (raw) {
            Object.assign(this, raw);
            this.links();
            this.screens = _.map(raw.screens, s => new Screen(s));

            this.created_at = new Date(this.created_at);
            this.updated_at = new Date(this.updated_at);
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
        if (!this.videoInfo) {
            if (this.isVideo() && this.meta) {
                let ffmpegProbe = JSON.parse(this.meta);
                this.videoInfo = new VideoCodecInfo(ffmpegProbe);
            } 
        }
        return this.videoInfo
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


class VideoFormat {
    bit_rate: number;
    duration: number;
    filename: string;
    format_long_name: string;
    format_name: string;
    nb_programs: number;
    nb_streams: number;
    probe_score: number;
    size: number;
    start_time: number;

    constructor(obj: any) {
        Object.assign(this, obj); // Lazy start...
    }
}

class VideoStream {
    avg_frame_rate: string;
    bit_rate: number;
    bits_per_raw_sample: number;
    chroma_location: string;
    closed_captions: number;
    codec_long_name: string;
    codec_name: string;
    codec_tag: string;
    codec_tag_string: string;
    codec_type: string;
    coded_height: number;
    coded_width: number;

    constructor(obj: any) {
        Object.assign(this, obj); // lazy
    }
}

// Represents some of the encoding that comes back from ffmpeg probe
export class VideoCodecInfo {
    format: VideoFormat;
    streams: Array<VideoStream>;

    public CanEncode: boolean = false;

    constructor(obj: any) {
        this.format = new VideoFormat(_.get(obj, "format"));
        this.streams = _.map(_.get(obj, "streams"), s => new VideoStream(s));

        this.CanEncode = this.getVideoCodecName() !== "hevc";
    }

    getVideoCodecName() {
        let streams = this.streams || [];
        for (let j = 0; j < streams.length; ++j) {
            let stream = streams[j];
            if (stream.codec_type == "video") {
                return stream.codec_name;  
            }
        }
    }
}