import * as _ from 'lodash';
import { Content, ContentData } from './content';
import { ApiDef } from './api_def';

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

export interface ContainerData {
  id?: number;
  name?: string;
  path?: string;
  total?: number;
  count?: number;
  contents?: ContentData[] | Content[] | null;
  previewUrl?: string;
  rowIdx?: number;
  loadState?: LoadStates;
  visible?: boolean;
}

export class Container {
  public contents: Array<Content> = [];
  public total: number = 0;
  public count: number = 0;
  public path: string = '';
  public name: string = '';
  public id: number = 0;
  public previewUrl: string = '';

  // Set on the initial content loads
  public loadState: LoadStates = LoadStates.NotLoaded;
  public visible: boolean = false;

  // All potential items that can be rendered from the contents
  public renderable: Array<Content> | undefined;
  public visibleSet: Array<Content> | undefined = [];

  // The currently selected Index
  public rowIdx: number = 0;

  constructor(cnt: ContainerData) {
    this.total = _.get(cnt, 'total') || 0;
    this.id = _.get(cnt, 'id') || 0;
    this.name = _.get(cnt, 'name') || '';
    this.path = _.get(cnt, 'path') || '';
    this.previewUrl = _.get(cnt, 'previewUrl') || '';
    this.rowIdx = _.get(cnt, 'rowIdx') || 0;
    this.loadState = _.get(cnt, 'loadState') || LoadStates.NotLoaded;
    this.visible = _.get(cnt, 'visible') || false;
    
    const contents = _.get(cnt, 'contents') || [];
    this.setContents(this.buildImgs(contents));
  }

  public getCurrentContent(): Content | null {
    let cntList = this.getContentList() || [];
    if (this.rowIdx >= 0 && this.rowIdx < cntList.length) {
      return cntList[this.rowIdx];
    }
    return cntList.length > 0 ? cntList[0] : null;
  }

  // For use in determining what should actually be visible at any time
  public getIntervalAround(currentItem: Content | null, requestedVisible: number = 4, before: number = 0): Content[] {
    this.visibleSet = [];

    let items = this.getContentList() || [];
    if (items.length === 0) {
      return [];
    }
    
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
      start = start >= 0 ? start : 0; // Make sure we don't go negative
    }
    this.visibleSet = items.slice(start, end) || [];
    return this.visibleSet;
  }

  public indexOf(item: Content, media: Array<Content> = []): number {
    const contents = media.length > 0 ? media : this.getContentList() || [];
    if (item && contents.length > 0) {
      return _.findIndex(contents, { id: item.id });
    }
    return -1;
  }

  public buildImgs(imgData: Array<ContentData | Content>): Content[] {
    return _.map(imgData, data => {
      return data instanceof Content ? data : new Content(data);
    });
  }

  public setContents(contents: Array<Content>): void {
    this.contents = _.sortBy(_.uniqBy(contents || [], 'id'), 'idx');
    this.count = this.contents.length;
    this.renderable = undefined;

    if (this.count === this.total) {
      this.loadState = LoadStates.Complete;
    } else if (this.loadState === LoadStates.Loading) {
      this.loadState = LoadStates.Partial;
    }
  }

  public addContents(contents: Array<Content>): Content[] {
    let sorted = _.sortBy((this.contents || []).concat(contents), 'idx');
    this.setContents(sorted);
    return sorted;
  }

  public getContent(rowIdx: number = 0): Content | null {
    const index = rowIdx === null ? this.rowIdx : rowIdx;
    if (index >= 0 && index < this.contents.length) {
      return this.contents[index];
    }
    return null;
  }

  // This is the actual URL you can get a pointer to for the scroll / load
  public getContentList(): Content[] {
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
