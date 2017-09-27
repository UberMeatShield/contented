import {OnInit, Component, EventEmitter, Input, Output, HostListener} from '@angular/core';
import {ContentedService, ApiDef} from './contented_service';

import * as _ from 'lodash';


class Directory {
    public path: string;
    public name: string;
    public contents: Array<string>;

    constructor(path, name, contents) {
        this.path = path || '';
        this.name = name || '';
        this.contents = contents || [];
    }

    public getContentList() {
        return _.map(this.contents, c => {
            let link = ApiDef.base + this.trail(this.path, '/') + this.trail(this.name, '/') + c || '';
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

    @Input() maxVisible: number = 2;
    @Input() idx: number = 0;
    constructor(public _contentedService: ContentedService) {

    }

    public fullScreen: boolean = false;
    public directories: Array<Directory>;
    public allD: Array<Directory>;

    @HostListener('document:keypress', ['$event'])
    public keyPress(evt: KeyboardEvent) {
        console.log("Keypress", evt);

        // Up (w)
        // down (s)
        // Left (a)
        switch (evt.key) {
            case 'a':
                this.prev();
                break;
            case 'd':
                this.next();
                break;
            case ' ':
                this.fullScreen = true;
                break;
            case 'q':
                this.fullScreen = false;
                break;
            case 'f':
                let visible = this.getVisibleDirectories();
                this.fullLoadDir(visible[0]);
                break;
            default:
                break;
        }
    }


    public ngOnInit() {
        console.log("Contented comp is alive.");
        this.loadDirs();
    }

    public loadDirs() {
        this._contentedService.getPreview().subscribe(
            res => { this.previewResults(res); },
            console.error
        );
    }

    public fullLoadDir(dir: Directory) {
        this._contentedService.getFullDirectory(dir.name).subscribe(
            res => { dir.contents = _.get(res, 'results'); },
            err => { console.error(err); }
        );
    }

    public reset() {
        this.idx = 0;
        this.allD = [];
    }

    public getVisibleDirectories() {
        if (this.allD) {
            let start = this.idx < this.allD.length ? this.idx : this.allD.length - 1;
            let end = start + this.maxVisible <= this.allD.length ? start + this.maxVisible : this.allD.length;
            return this.allD.slice(start, end);
        }
        return [];
    }

    public next() {
        if (this.allD && this.idx + 1 < this.allD.length) {
            this.idx++;
        }
    }

    public prev() {
        if (this.idx > 0) {
            this.idx--;
        }
    }

    public previewResults(response) {
        console.log("Results returned from the preview results.", response);
        let path = _.get(response, 'path');

        this.allD = _.map(_.get(response, 'results') || [], (contents, dir) => {
            return new Directory(path, dir, contents);
        });
    }
}
