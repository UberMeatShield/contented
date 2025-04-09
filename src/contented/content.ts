import * as _ from 'lodash';
import { ApiDef } from './api_def';
import { Screen, ScreenSchema } from './screen';

import { z } from 'zod';

// Why does a TAG have an id?!?!  Because goBuffalo really likes the id field.
export const TagSchema = z.object({
  id: z.string(),
  tag_type: z.string().optional(),
});
export type TagInterface = z.infer<typeof TagSchema>;
export class Tag implements TagInterface {
  id: string = '';
  tag_type: string = '';

  constructor(data: Partial<TagInterface> = {}) {
    this.update(data);
  }

  update(data: Partial<TagInterface> = {}) {
    const s = TagSchema.parse(data);
    Object.assign(this, s);
  }

  isProblem() {
    let arr = this.id ? this.id.split(' ') : [this.id];
    if (arr?.length > 1) {
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

/**
 * I really don't like the duplication but the Zod Class implementation bails on some of this pretty hard
 */
export const VideoFormatSchema = z.object({
  bit_rate: z.coerce.number(),
  duration: z.coerce.number(),
  filename: z.string(),
  format_long_name: z.string(),
  format_name: z.string(),
  nb_programs: z.coerce.number(),
  nb_streams: z.coerce.number(),
  probe_score: z.coerce.number(),
  size: z.coerce.number(),
  start_time: z.coerce.number(),
});

export type VideoFormatInterface = z.infer<typeof VideoFormatSchema>;
export class VideoFormat implements VideoFormatInterface {
  bit_rate: number = 0;
  duration: number = 0;
  filename: string = '';
  format_long_name: string = '';
  format_name: string = '';
  nb_programs: number = 0;
  nb_streams: number = 0;
  probe_score: number = 0;
  size: number = 0;
  start_time: number = 0;

  constructor(data: any = {}) {
    this.update(data);
  }

  update(data: any = {}) {
    const s = VideoFormatSchema.parse(data);
    Object.assign(this, s);
  }

  get durationSeconds(): number {
    if (!isNaN(this.duration)) {
      return Math.floor(this.duration);
    }
    return 0;
  }
}

export const VideoStreamSchema = z.object({
  avg_frame_rate: z.string(),
  bit_rate: z.coerce.number(),
  bits_per_raw_sample: z.coerce.number().optional(),
  chroma_location: z.string().optional(),
  closed_captions: z.coerce.number().optional(),
  codec_long_name: z.string(),
  codec_name: z.string(),
  codec_tag: z.string().optional(),
  codec_tag_string: z.string().optional(),
  codec_type: z.string().optional(),
  coded_height: z.coerce.number().optional(),
  coded_width: z.coerce.number().optional(),
  duration: z.coerce.number(),
});

export type VideoStreamInterface = z.infer<typeof VideoStreamSchema>;
export class VideoStream implements VideoStreamInterface {
  avg_frame_rate: string = '';
  bit_rate: number = 0;
  bits_per_raw_sample: number = 0;
  chroma_location: string = '';
  closed_captions: number = 0;
  codec_long_name: string = '';
  codec_name: string = '';
  codec_tag: string = '';
  codec_tag_string: string = '';
  codec_type: string = '';
  coded_height: number = 0;
  coded_width: number = 0;
  duration: number = 0;

  constructor(data: any = {}) {
    this.update(data);
  }

  update(data: any = {}) {
    const s = VideoStreamSchema.parse(data);
    Object.assign(this, s);
  }
}

// Represents some of the encoding that comes back from ffmpeg probe
export const VideoCodecInfoSchema = z.object({
  format: VideoFormatSchema.optional(),
  streams: VideoStreamSchema.array().optional(),
});
export type VideoCodecInfoInterface = z.infer<typeof VideoCodecInfoSchema>;

export class VideoCodecInfo implements VideoCodecInfoInterface {
  format?: VideoFormat | undefined;
  streams?: VideoStreamInterface[];

  constructor(data: Partial<VideoCodecInfoInterface> = {}) {
    this.update(data);
  }

  update(data: Partial<VideoCodecInfoInterface> = {}) {
    const s = VideoCodecInfoSchema.parse(data);
    Object.assign(this, s);

    if (s.format) {
      this.format = new VideoFormat(s.format);
    }

    this.streams = (s.streams || []).map(stream => new VideoStream(stream));
  }

  get CanEncode() {
    return this.getVideoCodecName() !== 'hevc';
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

export const ContentSchema = z.object({
  id: z.number(),
  src: z.string(),
  preview: z.string().optional(), // Name of the preview, if not set we do not have one.
  idx: z.number().default(0).optional(),
  description: z.string().default('').optional(),

  content_type: z.string().optional(),
  container_id: z.number().optional(),
  size: z.number().default(0).optional(), // size in bytes

  // Only defined currently on video
  encoding: z.string().optional(),
  screens: ScreenSchema.array().default([]).optional(),
  tags: TagSchema.array().default([]).optional(),
  meta: z.string().optional(),
  fullText: z.string().default('').optional(),

  created_at: z.string().optional(),
  updated_at: z.string().optional(),
  duplicate: z.boolean().default(false),
});

export type ContentInterface = z.infer<typeof ContentSchema>;

export class Content implements ContentInterface {
  id: number = 0;
  src: string = '';
  preview: string = '';
  idx: number = 0;
  description: string = '';
  content_type: string = '';
  container_id: number = 0;
  size: number = 0;
  encoding: string = '';
  screens: Screen[] = [];
  tags: Tag[] = [];
  meta: string = '';
  fullText: string = '';
  created_at: string = '';
  updated_at: string = '';
  duplicate: boolean = false;
  videoInfoParsed: VideoCodecInfo | undefined = undefined;

  constructor(data: any = {}) {
    this.update(data);
  }

  update(data: any = {}) {
    const s = ContentSchema.parse(data);
    Object.assign(this, s);

    this.screens = (this.screens || []).map(screen => new Screen(screen));
    this.tags = (this.tags || []).map(tag => new Tag(tag));

    if (this.isVideo() && this.meta) {
      this.videoInfoParsed = this.getVideoInfo();
    }
  }

  get createdAt() {
    if (this.created_at) {
      return new Date(this.created_at);
    }
    return undefined;
  }

  get updatedAt() {
    if (this.updated_at) {
      return new Date(this.updated_at);
    }
  }

  get previewUrl() {
    return `${ApiDef.contented.preview}${this.id}`;
  }

  get fullUrl() {
    return `${ApiDef.contented.view}${this.id}`;
  }

  get videoInfo(): VideoCodecInfo | undefined {
    return this.getVideoInfo();
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
    if (!this.videoInfoParsed) {
      if (this.isVideo() && this.meta) {
        try {
          let ffmpegProbe = JSON.parse(this.meta);
          this.videoInfoParsed = new VideoCodecInfo(ffmpegProbe);
        } catch (e) {
          console.error('Failed to parse video meta for ', this.id, e);
        }
      }
    }
    return this.videoInfoParsed;
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
}
