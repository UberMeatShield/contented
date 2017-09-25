import {Injectable} from '@angular/core';
import {Http, URLSearchParams, Request, RequestOptions, Headers} from '@angular/http';
import {Observable} from 'rxjs';

import * as _ from 'lodash';
let base = window.location.origin + '/';
export let ApiDef = {
    contented: {
        preview: base + '/content/',
        fulldir: base + '/content/{dir}'
    }
};


@Injectable()
export class ContentedService {

    public options: RequestOptions;
    constructor(private http: Http) {
        let headers = new Headers({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        });
        this.options = new RequestOptions({ headers: headers });
    }

    public getPreview() {
        return this.http.get(ApiDef.contented.preview, this.options)
          .map(res => res.json())
          .catch(err => this.handleError(err));
    }

    public getFullDirectory(dir: string) {
        let url = ApiDef.contented.fulldir.replace('{dir}', dir);
        return this.http.get(url, this.options)
          .map(res => res.json())
          .catch(err => this.handleError(err));
    }

    public handleError(err: any) {
        console.error("Failed to load this call", err);
        let parsed = err;
        if (err && err._body) {
            try {
                parsed = JSON.parse(err._body);
                parsed['status'] = err.status;
                if (_.isEmpty(parsed)) {
                    parsed = {
                        error: "No actual response data or response was somehow empty.",
                        debug: err._body,
                        status: err.status
                    };
                }
            } catch (e) {
                parsed = {
                    "error": !_.isEmpty(err._body) ? "No JSON was returned in this error" : "Unknown error has occurred.",
                    "status": err.status,
                    "debug": err._body
                };
            }
        }
        return Observable.fromPromise(Promise.reject(parsed));
    }
}
