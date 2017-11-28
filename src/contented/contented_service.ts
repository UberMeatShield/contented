import {Injectable} from '@angular/core';
import {HttpClient, HttpHeaders, HttpParams, HttpErrorResponse} from '@angular/common/http';
import {ApiDef} from './api_def';

// The manner in which RxJS does this is really stupid, saving 50K for hours of dev time is fail
import {Observable} from 'rxjs/Observable';
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/catch';
import 'rxjs/add/observable/fromPromise';
import 'rxjs/add/observable/forkJoin';
import 'rxjs/add/operator/finally';
import 'rxjs/add/operator/debounceTime';
import 'rxjs/add/operator/distinctUntilChanged';

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
          .catch(err => this.handleError(err));
    }

    public getFullDirectory(dir: string) {
        let url = ApiDef.contented.fulldir.replace('{dir}', dir);
        return this.http.get(url, this.options)
          .catch(err => this.handleError(err));
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
        return Observable.fromPromise(Promise.reject(parsed));
    }
}
