import {Injectable} from '@angular/core';
import {HttpClient, HttpHeaders, HttpParams, HttpErrorResponse} from '@angular/common/http';
import {Container, LoadStates} from './container';
import {Media} from './media';
import {ApiDef} from './api_def';

// The manner in which RxJS does this is really stupid, saving 50K for hours of dev time is fail
import {Observable, forkJoin, from as observableFrom} from 'rxjs';
import {catchError, map, finalize} from 'rxjs/operators';

import * as _ from 'lodash';
@Injectable()
export class ContentedService {

    public options = null;
    public LIMIT = 5000; // Default limit will use the server limit in the query
    // public LIMIT = 1; // Default limit will use the server limit in the query

    constructor(private http: HttpClient) {
        let headers = new HttpHeaders({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        });
        this.options = {headers: headers};
    }

    public getPreview() {
        return this.http.get(ApiDef.contented.containers, this.options)
            .pipe(
                map(res => {
                    return _.map(res, cnt => new Container(cnt));
                }),
                catchError(err => this.handleError(err))
            );
    }

    // Do a preview load (should it be API?)

    // TODO: Make all the test mock data new and or recent
    public download(cnt: Container, rowIdx: number) {
        console.log("Attempting to download", cnt, rowIdx);

        let img: Media = cnt.contents[rowIdx];
        let filename = cnt && rowIdx >= 0 && rowIdx < cnt.contents.length ? cnt.contents[rowIdx].src : '';
        if (!filename) {
            console.log("No file specified at rowIdx", rowIdx);
        }
        let downloadUrl = ApiDef.contented.download.replace('{mcID}', img.id);
        console.log("DownloadURL", downloadUrl);
        window.open(downloadUrl);
    }

    public fullLoadDir(cnt, limit = null) {
        if (cnt.count === cnt.total) {
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
                    this.getFullContainer(cnt.id, offset, limit).subscribe(res => {
                        _.delay(() => {
                            cnt.addContents(cnt.buildImgs(res));
                            yupResolve(cnt);
                        }, idx * 500);
                    }, err => {
                        console.error('Failed to load', err);
                        nopeReject(err);
                    });
                });

                if (calls.length > 30) { // TODO: Make something else sensible here.
                    break;
                }
                calls.push(observableFrom(delayP));
            }

            // Join all the results and let the call function resolve once the cnt is updated.
            return forkJoin(calls).pipe().subscribe(
                results => {
                    resolve(cnt);
                },
                err => {
                    console.error('Could not load all results', err);
                    reject(err);
                }
            );
        });
        return observableFrom(p);
    }

    public loadMoreInDir(cnt: Container, limit = null) {
        return this.getFullContainer(cnt.id, cnt.count, limit);
    }

    public getFullContainer(cnt: string, offset: number = 0, limit: number = null) {
        let url = ApiDef.contented.media.replace('{cId}', cnt);
        return this.http.get(url, {
            params: this.getPaginationParams(offset, limit),
            headers: this.options.headers
        }).pipe(catchError(err => this.handleError(err)));
    }

    public getPaginationParams(offset: number = 0, limit: number = 0) {
        if (limit <= 0 || limit == null) {
            limit = this.LIMIT;
        }
        let params = new HttpParams()
          .set('page', '' + (Math.floor(offset / limit) + 1))
          .set('per_page', '' + limit);
        return params;
    }


    // TODO: Create a pagination page for offset limit calculations
    public initialLoad(cnt: Container) {
        if (cnt.loadState === LoadStates.NotLoaded) {
            cnt.loadState = LoadStates.Loading;

            let url = ApiDef.contented.media.replace('{cId}', cnt.id);
            return this.http.get(url, {
                params: this.getPaginationParams(0, this.LIMIT),
                headers: this.options.headers
            }).pipe(map((imgData: Array<any>) => {
                return cnt.addContents(cnt.buildImgs(imgData));
            }));
        }
    }

    public searchMedia(text: string, offset: number = 0, limit: number = 0) {
        let params = this.getPaginationParams(offset, limit);
        params = params.set("text", text);

        return this.http.get(ApiDef.contented.search, {
            params: params
        });
    }

    public handleError(err: HttpErrorResponse) {
        console.error("Failed to handle API call error", err);
        let parsed = {};
        if (_.isObject(err.error)) {
            parsed = _.clone(err.error);
        } else {
           try {
                parsed = JSON.parse(err.error) || {};
                if (_.isEmpty(parsed)) {
                    parsed = {error: "No response error, are you logged in?", debug: err.error};
                }
            } catch (e) {
                console.error("Failed to parse the json result from the API call.");
                parsed = {
                    error: !_.isEmpty(err.error) ? "Exception, non json returned." : "Unhandled exception on the server.",
                    debug: err.error
                };
            }
        }
        if (_.isEmpty(parsed)) {
            parsed = {error: 'Unknown error, or no error text in the result?'};
        }
        parsed['url'] = err.url;
        parsed['code'] = err.status;
        return observableFrom(Promise.reject(parsed));
    }
}
