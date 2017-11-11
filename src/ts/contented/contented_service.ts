import {Injectable} from '@angular/core';
import {HttpClient, HttpHeaders, HttpParams, HttpErrorResponse} from '@angular/common/http';
import {Observable} from 'rxjs';

import * as _ from 'lodash';
let base = window.location.origin + '/';
export let ApiDef = {
    base: base,
    contented: {
        preview: base + 'content/',
        fulldir: base + 'content/{dir}'
    }
};


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
