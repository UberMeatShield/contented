import {Injectable} from '@angular/core';
import {HttpClient, HttpHeaders, HttpParams, HttpErrorResponse} from '@angular/common/http';
import {Directory} from './directory';
import {ApiDef} from './api_def';

// The manner in which RxJS does this is really stupid, saving 50K for hours of dev time is fail
import {Observable, from as observableFrom} from 'rxjs';
import {catchError, map, finalize} from 'rxjs/operators';

import * as _ from 'lodash';
@Injectable()
export class ContentedService {

    public options = null;
    constructor(private http: HttpClient) {
        let headers = new HttpHeaders({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        });
        this.options = {headers: headers};
    }

    public getPreview() {
        return this.http.get(ApiDef.contented.preview, this.options)
          .pipe(catchError(err => this.handleError(err)));
    }

    public download(dir: Directory, rowIdx: number) {
        console.log("Attempting to download", dir, rowIdx);

        let filename = dir && rowIdx >= 0 && rowIdx < dir.contents.length ? dir.contents[rowIdx] : '';
        if (!filename) {
            console.log("No file specified, wtf", rowIdx);
        }
        let downloadUrl = ApiDef.contented.download.replace('{id}', dir.id).replace('{filename}', filename);
        console.log("DownloadURL", downloadUrl);
        window.open(downloadUrl);
    }

    public getFullDirectory(dir: string) {
        let url = ApiDef.contented.fulldir.replace('{dir}', dir);
        return this.http.get(url, this.options)
          .pipe(catchError(err => this.handleError(err)));
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
