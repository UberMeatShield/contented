import {OnInit, Component, EventEmitter, Input, Output} from '@angular/core';
import {ContentedService, ApiDef} from './contented_service';

import * as _ from 'lodash';


class Directory {
    public path: string;
    public dir: string;
    public contents: Array<string>;

    constructor(path, dir, contents) {
        this.path = path || '';
        this.dir = dir || '';
        this.contents = contents || [];
    }

    public getContentList() {
        return _.map(this.contents, c => {
            let link = ApiDef.base + this.trail(this.path, '/') + this.trail(this.dir, '/') + c || '';
            console.log("Get contents list", link);
            return link;
        });
    }

    public trail(path: string, whatWith: string) {
        if (path[path.length - 1] !== whatWith) {
            return path + whatWith;
        }
        return path;
    }
}


@Component({
    selector: 'contented-main',
    template: require('./contented.ng.html')
})
export class ContentedCmp implements OnInit {

    constructor(public _contentedService: ContentedService) {

    }

    public directories: Array<Directory>;
    public ngOnInit() {
        console.log("Contented comp is alive.");
        this._contentedService.getPreview().subscribe(
            res => { this.previewResults(res); },
            console.error
        );
    }

    public previewResults(response) {
        console.log("Results returned from the preview results.", response);
        let path = _.get(response, 'path');

        this.directories = _.map(_.get(response, 'results') || [], (contents, dir) => {
            return new Directory(path, dir, contents);
        });
    }
}
