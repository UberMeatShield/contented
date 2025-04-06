import * as _ from 'lodash';
import { ApiDef } from './api_def';
import { Screen, ScreenData } from './screen';

export interface TagData {
  id?: string;
  tag_type?: string;
}

// Why does a TAG have an id?!?!  Because goBuffalo really likes the id field.
export class Tag {
  public id: string = '';
  public tag_type: string = '';

  constructor(obj: TagData | string) {
    if (typeof obj === 'string') {
      this.id = obj;
    } else if (obj) {
      this.id = obj.id || '';
      this.tag_type = obj.tag_type || '';
    }
  }

  isProblem() {
    let arr = this.id ? this.id.split(' ') : [this.id];
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

export interface ContentData {
  id?: number;
  src?: string;
  preview?: string;
  idx?: number;
  description?: string;
  content_type?: string;
  container_id?: number;
  size?: number;
  encoding?: string;
  previewUrl?: string;
  fullUrl?: string;
  screens?: Array<ScreenData> | null;
  tags?: Array<TagData | string> | null;
  meta?: string;
  fullText?: string;
  created_at?: string | Date;
  updated_at?: string | Date;
  duplicate?: boolean;
}

export class Content {
  public id: number = 0;
  public src: string = '';
  public preview: string = ''; // Name of the preview, if not set we do not have one.
  public idx: number = 0;
  public description: string = '';

  // Awkward that buffalo makes the API use container_id like the DB
  // side of things and in url params by default.  So I guess mixed
  // cases it is.
  public content_type: string = '';
  public container_id: number = 0;
  public size: number = 0; // size in bytes

  // Only defined currently on video
  public encoding: string | undefined = undefined;
  public previewUrl: string = '';
  public fullUrl: string = '';
  public screens: Array<Screen> = [];
  public tags: Array<Tag> = [];
  public meta: string = '';

  public fullText: string | undefined = undefined;
  public videoInfo?: VideoCodecInfo;

  public created_at: Date | undefined; 
  public updated_at: Date | undefined;

  public duplicate: boolean = false;

  constructor(obj: ContentData = {}) {
    this.fromJson(obj);
  }

  public fromJson(raw: ContentData) {
    if (raw) {
      this.update(raw);
    }
  }

  public update(raw: ContentData) {
    Object.assign(this, _.omit(raw, ['screens', 'tags', 'created_at', 'updated_at']));
    this.links();
    
    this.screens = _.map(raw.screens || [], s => new Screen(s));
    this.tags = _.map(raw.tags || [], t => new Tag(t));

    if (raw.created_at) {
      this.created_at = raw.created_at instanceof Date ? raw.created_at : new Date(raw.created_at);
    }
    
    if (raw.updated_at) {
      this.updated_at = raw.updated_at instanceof Date ? raw.updated_at : new Date(raw.updated_at);
    }

    if (this.isVideo()) {
      this.getVideoInfo();
    }
  }

  public isImage() {
    return this.content_type ? !!this.content_type.match('image') : false;
  }

  public isVideo() {
    return this.content_type ? !!this.content_type.match('video') : false;
  }

  public isText() {
    return this.content_type ? !!this.content_type.match('text') : false;
  }

  getVideoInfo() {
    if (!this.videoInfo) {
      if (this.isVideo() && this.meta) {
        try {
          let ffmpegProbe = JSON.parse(this.meta);
          this.videoInfo = new VideoCodecInfo(ffmpegProbe);
        } catch (e) {
          console.error('Failed to parse video meta for ', this.id, e);
        }
      }
    }
    return this.videoInfo;
  }

  // Images will just work as a preview source, but video (with no preview) and
  // text and zips etc should use a content_type based style preview.  This prevents
  // broken image links when no previews are found for these types.
  public shouldUseTypedPreview() {
    if (_.isEmpty(this.preview)) {
      if (this.isImage()) {
        return ''; // image we can just display the image (maybe large)
      } else if (this.isVideo()) {
        return 'videocam'; // material icon
      } else if (this.isText()) {
        return 'article'; // material icon
      }
    }
    return '';
  }

  public links() {
    this.previewUrl = `${ApiDef.contented.preview}${this.id}`;
    this.fullUrl = `${ApiDef.contented.view}${this.id}`;
  }
}

export interface VideoFormatData {
  bit_rate?: number;
  duration?: number;
  filename?: string;
  format_long_name?: string;
  format_name?: string;
  nb_programs?: number;
  nb_streams?: number;
  probe_score?: number;
  size?: number;
  start_time?: number;
}

class VideoFormat {
  bit_rate?: number;
  duration?: number;
  filename: string = '';
  format_long_name?: string;
  format_name?: string;
  nb_programs?: number;
  nb_streams?: number;
  probe_score?: number;
  size?: number;
  start_time?: number;

  constructor(obj: VideoFormatData) {
    if (obj) {
      Object.assign(this, obj);

      if (this.duration && !isNaN(this.duration)) {
        this.duration = Math.floor(this.duration);
      }
    }
  }
}

export interface VideoStreamData {
  avg_frame_rate?: string;
  bit_rate?: number;
  bits_per_raw_sample?: number;
  chroma_location?: string;
  closed_captions?: number;
  codec_long_name?: string;
  codec_name?: string;
  codec_tag?: string;
  codec_tag_string?: string;
  codec_type?: string;
  coded_height?: number;
  coded_width?: number;
  duration?: number;
}

class VideoStream {
  avg_frame_rate?: string;
  bit_rate?: number;
  bits_per_raw_sample?: number;
  chroma_location?: string;
  closed_captions?: number;
  codec_long_name?: string;
  codec_name?: string;
  codec_tag?: string;
  codec_tag_string?: string;
  codec_type?: string;
  coded_height?: number;
  coded_width?: number;
  duration?: number;

  constructor(obj: VideoStreamData) {
    if (obj) {
      Object.assign(this, obj);
    }
  }
}

export interface VideoCodecInfoData {
  format?: VideoFormatData;
  streams?: VideoStreamData[];
}

// Represents some of the encoding that comes back from ffmpeg probe
export class VideoCodecInfo {
  format: VideoFormat;
  streams: Array<VideoStream>;

  public CanEncode: boolean = false;

  constructor(obj: VideoCodecInfoData) {
    this.format = new VideoFormat(_.get(obj, 'format') || {});
    this.streams = _.map(_.get(obj, 'streams') || [], s => new VideoStream(s));
    this.CanEncode = this.getVideoCodecName() !== 'hevc';
  }

  getVideoStream() {
    return (this.streams || []).find(stream => stream.codec_type === 'video');
  }

  getResolution() {
    const stream = this.getVideoStream();
    if (stream) {
      return stream.coded_width + 'x' + stream.coded_height;
    }
    return '';
  }

  getVideoCodecName() {
    const stream = this.getVideoStream();
    return stream?.codec_name || '';
  }
}
