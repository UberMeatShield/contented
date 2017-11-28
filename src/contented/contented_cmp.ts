import {OnInit, Component, EventEmitter, Input, Output, HostListener} from '@angular/core';
import {ContentedService} from './contented_service';
import {Directory} from './directory';

import * as _ from 'lodash';

@Component({
    selector: 'contented-main',
    templateUrl: 'contented.ng.html'
})
export class ContentedCmp implements OnInit {

    @Input() maxVisible: number = 2; // How many of the loaded directories should we be viewing
    @Input() rowIdx: number = 0; // Which row (directory) are we in
    @Input() idx: number = 0; // Which item within the directory are we viewing

    public loading: boolean = false;
    public previewWidth: number; // Based on current client page sizes, scale the preview images natually
    public previewHeight: number; // height for the previews ^

    private currentViewItem: string; // The current indexed item that is considered selected
    public fullScreen: boolean = true; // Should we view fullscreen the current item
    public directories: Array<Directory>; // Current set of visible directories
    public allD: Array<Directory>; // All the directories we have loaded

    constructor(public _contentedService: ContentedService) {
        this.calculateDimensions();
    }

    @HostListener('document:keypress', ['$event'])
    public keyPress(evt: KeyboardEvent) {
        console.log("Keypress", evt);

        // Up (w)
        // down (s)
        // Left (a)
        switch (evt.key) {
            case 'w':
                this.prev();
                break;
            case 's':
                this.next();
                break;
            case 'a':
                this.rowPrev();
                break;
            case 'd':
                this.rowNext();
                break;
            case 'e':
                this.viewFullscreen();
                break;
            case 'q':
                this.hideFullscreen();
                break;
            case 'f':
                this.fullLoad();
                break;
            default:
                break;
        }
    }

    public fullLoad() {
        let visible = this.getVisibleDirectories();
        this.fullLoadDir(visible[0]);
    }

    public viewFullscreen() {
        this.currentViewItem = this.getCurrentLocation();
        this.fullScreen = true;
    }

    public hideFullscreen() {
        this.currentViewItem = null;
        this.fullScreen = false;
    }

    public ngOnInit() {
        console.log("Contented comp is alive.");
        this.loadDirs();
    }

    public loadDirs() {
        this._contentedService.getPreview().subscribe(
            res => { this.previewResults(res); },
            err => { console.error(err); }
        );
    }

    public fullLoadDir(dir: Directory) {
        if (dir.count < dir.total) {
            this.loading = true;
            this._contentedService.getFullDirectory(dir.id)
                .finally(() => {this.loading = false; })
                .subscribe(
                    res => { this.dirResults(dir, res); },
                    err => { console.error(err); }
                );
        }
    }

    public dirResults(dir: Directory, response) {
        console.log("Full Directory loading, what is in the results?", response);
        dir.setContents(_.get(response, 'contents'));
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

    public rowNext() {
        let dirs = this.getVisibleDirectories();
        this.idx = 0;
        if (!_.isEmpty(dirs)) {
            let items = dirs[0].getContentList();
            if (!_.isEmpty(items) && this.rowIdx < items.length) {
                this.rowIdx++;
                this.currentViewItem = this.getCurrentLocation();
            }
        }
    }

    public rowPrev() {
        this.idx = 0;
        if (this.rowIdx > 0) {
            this.rowIdx--;
            this.currentViewItem = this.getCurrentLocation();
        }
    }

    public next() {
        if (this.allD && this.idx + 1 < this.allD.length) {
            this.idx++;
            this.rowIdx = 0;
        }
    }

    public prev() {
        if (this.idx > 0) {
            this.idx--;
            this.rowIdx = 0;
        }
    }

    public getCurrentDir() {
        let dirs = this.getVisibleDirectories();
        if (!_.isEmpty(dirs)) {
            return dirs[0];
        }
        return null;
    }

    public imgLoaded(evt) {
        let img = evt.target;
        console.log("Img Loaded", img.naturalHeight, img.naturalWidth, img);
    }

    public getCurrentLocation() {
        let dir = this.getCurrentDir();
        if (dir && !_.isEmpty(dir.getContentList())) {
            let contentList = dir.getContentList();
            if (this.rowIdx >= 0 && this.rowIdx < contentList.length) {
                return contentList[this.rowIdx];
            }
        }
    }

    // TODO: Being called abusively in the directive rather than on page resize events
    public calculateDimensions() {
        let width = !window['jasmine'] ? document.body.clientWidth : 800;
        let height = !window['jasmine'] ? document.body.clientHeight : 800;

        this.previewWidth = (width / 4) - 20;
        this.previewHeight = (height / this.maxVisible) - 20;
    }

    public previewResults(response) {
        console.log("Results returned from the preview results.", response);
        this.allD = _.map(_.get(response, 'results') || [], dir => {
            return new Directory(dir);
        });
    }
}

