import * as _ from 'lodash';
import { Content, ContentSchema } from './content';
import { ApiDef } from './api_def';
import { z } from 'zod';

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
  Complete,
}

export const ContainerSchema = z.object({
  id: z.number(),
  name: z.string().optional(),
  previewUrl: z.string().optional(),
  contents: z.array(ContentSchema).nullable().optional(),
  total: z.number().default(0),
  description: z.string().default('').optional(),
});

export type IContainer = z.infer<typeof ContainerSchema>;

export class Container implements IContainer {
  public contents: Array<Content> = [];
  public total: number;
  public count: number;
  public path: string;
  public name: string;
  public id: number;
  public previewUrl: string;
  public description: string;

  // Set on the initial content loads
  public loadState: LoadStates = LoadStates.NotLoaded;
  public visible: boolean = false;

  // All potential items that can be rendered from the contents
  public renderable: Array<Content>;
  public visibleSet: Array<Content> = [];

  // The currently selected Index
  public rowIdx: number = 0;

  constructor(cnt: any) {
    this.update(cnt);
  }

  public update(cnt: any) {
    const c = ContainerSchema.parse(cnt);
    Object.assign(this, c);

    const contents = cnt?.contents ? cnt.contents.map(mc => new Content(mc)) : [];
    this.setContents(contents);
  }

  public getCurrentContent() {
    let cntList = this.getContentList() || [];
    if (this.rowIdx >= 0 && this.rowIdx < cntList.length) {
      return cntList[this.rowIdx];
    }
    return cntList[0];
  }

  // For use in determining what should actually be visible at any time
  public getIntervalAround(currentItem: Content, requestedVisible: number = 4, before: number = 0) {
    this.visibleSet = [];

    let items = this.getContentList() || [];
    let start = 0;
    let max = requestedVisible < items.length ? requestedVisible : items.length;

    // Need to look it up by ID
    if (currentItem) {
      start = this.indexOf(currentItem);
      start = start >= 0 ? start : 0;
      start = before && start - before > 0 ? start - before : 0;
    }

    let end = start + (max >= 1 ? max : 4);
    end = end < items.length ? end : items.length;
    let interval = end - start;
    if (interval < max) {
      start = start - (max - interval);
    }
    this.visibleSet = items.slice(start, end) || [];
    return this.visibleSet;
  }

  public indexOf(item: Content, contents?: Array<Content>) {
    contents = contents || this.getContentList() || [];
    if (item && contents) {
      return _.findIndex(contents, { id: item.id });
    }
    return -1;
  }

  public setContents(contents: Array<Content>) {
    this.contents = _.sortBy(_.uniqBy(contents || [], 'id'), 'idx');
    this.count = this.contents.length;
    this.renderable = [];

    if (this.count === this.total) {
      this.loadState = LoadStates.Complete;
    } else if (this.loadState === LoadStates.Loading) {
      this.loadState = LoadStates.Partial;
    }
  }

  public addContents(contents: Array<Content>) {
    if (!contents) {
      return;
    }
    if (!contents || !contents.forEach) {
      console.error('No contents to add', contents?.length);
      return;
    }

    contents.forEach(c => {
      if (!(c instanceof Content)) {
        throw new Error(`Content is not an instance of Content ${c}`);
      }
    });

    let sorted = _.sortBy((this.contents || []).concat(contents), 'idx');
    this.setContents(sorted);
    return sorted;
  }

  public getContent(rowIdx?: number) {
    rowIdx = rowIdx === undefined ? this.rowIdx : rowIdx;
    if (rowIdx >= 0 && rowIdx < this.contents.length) {
      return this.contents[rowIdx];
    }
    return undefined;
  }

  // This is the actual URL you can get a pointer to for the scroll / load
  public getContentList() {
    if (!this.renderable) {
      this.renderable = _.map(this.contents, (c: Content) => {
        return c;
      });
    }
    return this.renderable || [];
  }
}

// No loading in the DB at this point so this will work for just the UI development
let favoriteContainer: Container;
export function getFavorites() {
  if (!favoriteContainer) {
    favoriteContainer = new Container({
      id: -42,
      name: 'Favorites',
      previewUrl: '', // Find a local one and use that
      contents: [],
      total: 0,
      count: 0,
      rowIdx: 0,
    });
  }
  return favoriteContainer;
}
