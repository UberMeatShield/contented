import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders, HttpParams, HttpErrorResponse } from '@angular/common/http';
import { Container, LoadStates } from './container';
import { Content, Tag } from './content';
import { Screen } from './screen';
import { TaskRequest, TASK_STATES } from './task_request';
import { ApiDef } from './api_def';
import { TAGS_RESPONSE } from './tagging_syntax';

// The manner in which RxJS does this is really stupid, saving 50K for hours of dev time is fail
import { forkJoin, Observable, from as observableFrom } from 'rxjs';
import { catchError, map } from 'rxjs/operators';
import { GlobalBroadcast } from './global_message';

import * as _ from 'lodash';
import z from 'zod';
//import { Z } from 'zod-class';

export const DirectionEnum = z.enum(['asc', 'desc']);
export const SearchSchema = z.object({
  search: z.string().optional(), // Searches inside description
  offset: z.number().optional().default(0),
  limit: z.number().optional(),
  tags: z.string().array().optional(),
  order: DirectionEnum.optional(),
});

export const ContentSearchSchema = SearchSchema.extend({
  cId: z.string().optional().nullable(), // Container Id
  contentType: z.string().optional(),
  text: z.string().optional(), // An exact search on file name
  duplicate: z.boolean().optional(),
});

export const ContainerSearchSchema = SearchSchema.extend({
  name: z.string().optional(), // Exact search on container name
});

export type ContentSearch = z.infer<typeof ContentSearchSchema>;
export type ContainerSearch = z.infer<typeof ContainerSearchSchema>;

/* This should work but doesn't because of node vs web issues.
export class ContentSearch extends Z.class({
    ...ContentSearchSchema._def.shape(),
}) {}
*/
export const TaskStatusEnum = z.enum(['new', 'pending', 'in_progress', 'canceled', 'error', 'done', 'invalid', '']);
export type TaskStatus = z.infer<typeof TaskStatusEnum>;

export const TaskTypes = {
  ENCODING: 'video_encoding',
  SCREENS: 'screen_capture',
  WEBP: 'webp_from_screens',
  TAGGING: 'tag_content',
  DUPES: 'detect_duplicates',
} as const;

// Odd but works because of a strange constant hackery found in the zod forums.
export const TaskEnum = z.enum([TaskTypes.ENCODING, ...Object.values(TaskTypes)]);

export const TaskSearchSchema = z.object({
  id: z.string().optional().nullable(), // The task.ID
  offset: z.number().optional().default(0),
  limit: z.number().optional(),
  contentID: z.string().optional(),
  containerID: z.string().optional(),
  search: z.string().optional(),
  status: TaskStatusEnum.optional(),
  operation: TaskEnum.optional(),
});
export type TaskSearch = z.infer<typeof TaskSearchSchema>;

@Injectable()
export class ContentedService {
  public options = null;
  public LIMIT = 5000; // Default limit will use the server limit in the query
  // public LIMIT = 1; // Default limit will use the server limit in the query

  constructor(private http: HttpClient) {
    let headers = new HttpHeaders({
      'Content-Type': 'application/json',
      Accept: 'application/json',
    });
    this.options = { headers: headers };
  }

  public getContainers() {
    return this.http.get(ApiDef.contented.containers, this.options).pipe(
      map((res: any) => {
        return {
          total: res.total,
          results: _.map(res.results, cnt => new Container(cnt)),
        };
      }),
      catchError(err => this.handleError(err))
    );
  }

  public getScreens(contentID: string): Observable<{ total: number; results: Array<Screen> }> {
    let url = ApiDef.contented.contentScreens.replace('{mcID}', contentID);
    return this.http.get(url, this.options).pipe(
      map((res: any) => {
        return {
          total: res.total,
          results: _.map(res.results, s => new Screen(s)),
        };
      }),
      catchError(err => this.handleError(err))
    );
  }

  public clearScreens(contentID: string): Observable<Content> {
    let url = ApiDef.contented.contentScreens.replace('{mcID}', contentID);
    return this.http.delete(url, this.options).pipe(
      map((res: any) => {
        return new Content(res);
      }),
      catchError(err => this.handleError(err))
    );
  }

  public getContent(contentID: string) {
    let url = ApiDef.contented.content.replace('{id}', contentID);
    return this.http.get(url, this.options).pipe(
      map(mc => {
        return new Content(mc);
      }),
      catchError(err => this.handleError(err))
    );
  }

  public removeContent(contentID: string) {
    let url = ApiDef.contented.content.replace('{id}', contentID);
    return this.http.delete(url, this.options).pipe(catchError(err => this.handleError(err)));
  }

  // Do a preview load (should it be API?)

  // TODO: Make all the test mock data new and or recent
  public download(cnt: Container, rowIdx: number) {
    console.log('Attempting to download', cnt, rowIdx);

    let content: Content = cnt.contents[rowIdx];
    let filename = cnt && rowIdx >= 0 && rowIdx < cnt.contents.length ? cnt.contents[rowIdx].src : '';
    if (!filename) {
      console.log('No file specified at rowIdx', rowIdx);
    }
    let downloadUrl = ApiDef.contented.download.replace('{mcID}', content.id);
    console.log('DownloadURL', downloadUrl);
    window.open(downloadUrl);
  }

  public getTextContent(content: Content) {
    let downloadUrl = ApiDef.contented.download.replace('{mcID}', content.id);
    return this.http.get(downloadUrl, { responseType: 'text' });
  }

  public fullLoadDir(cnt, limit = null) {
    if (cnt.count === cnt.total) {
      console.log('Count = total, ignoring', cnt);
      return observableFrom(Promise.resolve(cnt));
    }

    limit = limit || this.LIMIT || 2000;
    // Build out a call to load all the possible data (all at once, it is fast)
    let p = new Promise((resolve, reject) => {
      let calls = [];
      let idx = 0;
      for (let offset = cnt.count; offset < cnt.total; offset += limit) {
        ++idx;
        let delayP = new Promise((yupResolve, nopeReject) => {
          this.getFullContainer(cnt.id, offset, limit).subscribe({
            next: res => {
              _.delay(() => {
                // Hmmm, buildImgs is strange and should be fixed up
                cnt.addContents(cnt.buildImgs(res.results));
                yupResolve(cnt);
              }, idx * 500);
            },
            error: err => {
              GlobalBroadcast.error('Failed to load', err);
              nopeReject(err);
            },
          });
        });

        if (calls.length > 30) {
          // TODO: Make something else sensible here.
          break;
        }
        calls.push(observableFrom(delayP));
      }

      // Join all the results and let the call function resolve once the cnt is updated.
      return forkJoin(calls)
        .pipe()
        .subscribe({
          next: results => {
            resolve(cnt);
          },
          error: err => {
            GlobalBroadcast.error('Could not load all results', err);
            reject(err);
          },
        });
    });
    return observableFrom(p);
  }

  public loadMoreInDir(cnt: Container, limit = null) {
    return this.getFullContainer(cnt.id, cnt.count, limit);
  }

  public getFullContainer(cnt: string, offset: number = 0, limit: number = null) {
    let url = ApiDef.contented.containerContent.replace('{cId}', cnt);
    return this.http
      .get(url, {
        params: this.getPaginationParams(offset, limit),
        headers: this.options.headers,
      })
      .pipe(
        map((res: any) => {
          return {
            total: res.total,
            results: _.map(res.results, c => new Content(c)),
          };
        }),
        catchError(err => this.handleError(err))
      );
  }

  public getPaginationParams(offset: number = 0, limit: number = 0) {
    if (limit <= 0 || limit == null) {
      limit = this.LIMIT;
    }
    let params = new HttpParams().set('page', '' + (Math.floor(offset / limit) + 1)).set('per_page', '' + limit);
    return params;
  }

  // TODO: Create a pagination page for offset limit calculations
  public initialLoad(cnt: Container) {
    if (cnt.loadState === LoadStates.NotLoaded) {
      cnt.loadState = LoadStates.Loading;

      let url = ApiDef.contented.containerContent.replace('{cId}', cnt.id);
      return this.http
        .get(url, {
          params: this.getPaginationParams(0, this.LIMIT),
          headers: this.options.headers,
        })
        .pipe(
          map((res: any) => {
            return cnt.addContents(cnt.buildImgs(res.results));
          })
        );
    }
  }

  public searchContainers(cntQ: ContainerSearch) {
    let params = this.getPaginationParams(cntQ.offset, cntQ.limit);
    if (cntQ.name) {
      params = params.set('text', cntQ.name);
    }
    if (cntQ.search) {
      params = params.set('search', cntQ.search);
    }
    if (cntQ.tags?.length > 0) {
      params = params.set('tags', JSON.stringify(cntQ.tags));
    }
    return this.http.get(ApiDef.contented.searchContainers, { params }).pipe(
      map((res: any) => {
        return {
          total: res.total,
          results: _.map(res.results, r => new Container(r)),
        };
      })
    );
  }

  // Could definitely use Zod here as a search type.  Maybe it is worth pulling in at this point.
  public searchContent(cs: ContentSearch) {
    let params = this.getPaginationParams(cs.offset, cs.limit);
    params = params.set('search', cs.text);
    if (cs.contentType) {
      params = params.set('contentType', cs.contentType);
    }
    if (cs.cId) {
      params = params.set('cId', cs.cId);
    }

    // GoBuffalo is being DUMB on the array parsing :(
    // params.get("tags[]") just returns the first entry if there are multiple
    if (cs.tags?.length > 0) {
      params = params.set('tags', JSON.stringify(cs.tags));
    }
    if (cs.duplicate) {
      params = params.set('duplicate', 'true');
    }
    return this.http
      .get(ApiDef.contented.searchContents, {
        params,
      })
      .pipe(
        map((res: any) => {
          return {
            total: res.total,
            results: _.map(res.results, r => new Content(r)),
          };
        })
      );
  }

  public saveContent(content: Content) {
    let url = ApiDef.contented.content.replace('{id}', content.id);
    return this.http.put(url, content).pipe(catchError(err => this.handleError(err)));
  }

  public handleError(err: HttpErrorResponse) {
    console.error('Error calling API', err);
    let parsed = {};
    if (_.isObject(err.error)) {
      parsed = _.clone(err.error);
    } else {
      try {
        parsed = JSON.parse(err.error) || {};
        if (_.isEmpty(parsed)) {
          parsed = {
            error: 'No response error, are you logged in?',
            debug: err.error,
          };
        }
      } catch (e) {
        console.error('Failed to parse the json result from the API call.');
        parsed = {
          error: !_.isEmpty(err.error) ? 'Exception, non json returned.' : 'Unhandled exception on the server.',
          debug: err.error,
        };
      }
    }
    if (_.isEmpty(parsed)) {
      parsed = { error: 'Unknown error, or no error text in the result?' };
    }
    parsed['url'] = err.url;
    parsed['code'] = err.status;
    return observableFrom(Promise.reject(parsed));
  }

  // This page allows server configuration of the home page display.
  splash() {
    return this.http.get(ApiDef.contented.splash).pipe(
      map((res: any) => {
        // Worth an actual class type?
        return {
          container: _.get(res, 'container.id') ? new Container(res.container) : null,
          content: _.get(res, 'content.id') ? new Content(res.content) : null,
          splashTitle: res.splashTitle || '',
          splashContent: res.splashContent || '',
          rendererType: res.rendererType || 'video',
        };
      })
    );
  }

  requestScreens(content: Content, count: number = 1, startTime: number = 2) {
    let url = ApiDef.contented.requestScreens.replace('{id}', content.id);
    url = url.replace('{count}', '' + count);
    url = url.replace('{startTimeSeconds}', '' + Math.floor(startTime));
    return this.http.post(url, {}).pipe(
      map(res => {
        return new TaskRequest(res);
      })
    );
  }

  encodeVideoContent(content: Content, codec: string = '') {
    let params = new HttpParams();
    params = params.set('codec', codec);
    let url = ApiDef.contented.encodeVideoContent.replace('{id}', content.id);
    return this.http.post(url, { params: params }).pipe(
      map(res => {
        return new TaskRequest(res);
      })
    );
  }

  // Determine what kinds of args we can provide
  createPreviewFromScreens(content: Content) {
    let url = ApiDef.contented.createPreviewFromScreens.replace('{id}', content.id);
    return this.http.post(url, {}).pipe(
      map(res => {
        return new TaskRequest(res);
      })
    );
  }

  // Determine what kinds of args we can provide
  createTagContentTask(content: Content) {
    let url = ApiDef.contented.createTagContentTask.replace('{id}', content.id);
    return this.http.post(url, {}).pipe(
      map(res => {
        return new TaskRequest(res);
      })
    );
  }

  findDuplicateForContentTask(content: Content) {
    let url = ApiDef.contented.contentDuplicatesTask.replace('{contentId}', content.id);
    return this.http.post(url, content).pipe(
      map(res => {
        return new TaskRequest(res);
      })
    );
  }

  containerDuplicatesTask(cnt: Container) {
    let url = ApiDef.contented.containerDuplicatesTask.replace('{containerId}', cnt.id);
    return this.http.post(url, cnt).pipe(
      map(res => {
        return [new TaskRequest(res)];
      })
    );
  }

  containerRemoveDuplicatesTask(cnt: Container) {
    let url = ApiDef.contented.containerRemoveDuplicatesTask.replace('{containerId}', cnt.id);
    return this.http.post(url, cnt).pipe(
      map(res => {
        return [new TaskRequest(res)];
      })
    );
  }

  containerPreviewsTask(cnt: Container, count: number = 16, startTimeSeconds: number = -1) {
    let url = ApiDef.contented.containerPreviewsTask.replace('{containerId}', cnt.id);
    url = url.replace('{count}', `${count}`).replace('{startTimeSeconds}', `${startTimeSeconds}`);
    return this.http.post(url, cnt).pipe(
      map(res => {
        console.log('Created container previews response', res);
        return _.map(res['results'], task => new TaskRequest(task));
      })
    );
  }

  containerVideoEncodingTask(cnt: Container) {
    let url = ApiDef.contented.containerVideoEncodingTask.replace('{containerId}', cnt.id);
    return this.http.post(url, cnt).pipe(
      map(res => {
        // Return an array of task requests I think
        console.log('Container Encoding task', res);
        return _.map(res['results'], task => new TaskRequest(task));
      })
    );
  }

  containerTaggingTask(cnt: Container) {
    let url = ApiDef.contented.containerTaggingTask.replace('{containerId}', cnt.id);
    return this.http.post(url, cnt).pipe(
      map(res => {
        return _.map(res['results'], task => new TaskRequest(task));
      })
    );
  }

  getTags(page: number = 1, perPage: number = 1000, tagType: string = '') {
    if (TAGS_RESPONSE.initialized) {
      return observableFrom(
        new Promise((resolve, reject) => {
          resolve(TAGS_RESPONSE);
        })
      );
    }
    let params = new HttpParams();
    params = params.set('page', '' + page);
    params = params.set('per_page', '' + perPage);
    params = params.set('tag_type', tagType);

    return this.http.get(ApiDef.contented.tags, { params: params }).pipe(
      map((res: any) => {
        return {
          total: res.total || 0,
          results: _.map(res.results, t => new Tag(t)),
        };
      })
    );
  }

  // TODO: Update this to a query object
  getTasks(query: TaskSearch) {
    // TODO: make a toParam() ?
    let params = this.getPaginationParams(query.offset, query.limit);
    if (query.id) {
      params = params.set('id', query.id);
    }
    if (query.contentID) {
      params = params.set('content_id', query.contentID);
    }
    if (query.containerID) {
      params = params.set('container_id', query.containerID);
    }
    if (query.status) {
      params = params.set('status', query.status);
    }
    if (query.search) {
      params = params.set('search', query.search);
    }
    return this.http.get(ApiDef.tasks.list, { params: params }).pipe(
      map((res: any) => {
        return {
          total: res.total,
          results: _.map(res.results, r => new TaskRequest(r)),
        };
      })
    );
  }

  cancelTask(task: TaskRequest) {
    const url = ApiDef.tasks.update.replace('{id}', task.id);
    const up = _.clone(task);
    up.status = TASK_STATES.CANCELED;
    return this.http.put(url, up).pipe(
      map((res: any) => {
        return new TaskRequest(res);
      })
    );
  }
}
