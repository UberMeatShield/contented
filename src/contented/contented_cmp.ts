import {OnInit, Component, EventEmitter, Input, Output, HostListener} from '@angular/core';
import {ContentedService} from './contented_service';
import {Container} from './container';
import {Media} from './media';
import {finalize, switchMap} from 'rxjs/operators';

import {ActivatedRoute, Router, ParamMap} from '@angular/router';
import {GlobalNavEvents, NavTypes} from './nav_events';

import * as _ from 'lodash';

@Component({
    selector: 'contented-main',
    templateUrl: 'contented.ng.html'
})
export class ContentedCmp implements OnInit {

    @Input() maxVisible: number = 2; // How many of the loaded containers should we be viewing
    @Input() rowIdx: number = 0; // Which row (media item) are we on
    @Input() idx: number = 0; // Which item within the container are we viewing

    public loading: boolean = false;
    public emptyMessage = null;
    public previewWidth: number = 200; // Based on current client page sizes, scale the preview images natually
    public previewHeight: number = 200; // height for the previews ^

    public currentViewItem: Media; // The current indexed item that is considered selected
    public currentDir: Container;
    public fullScreen: boolean = false; // Should we view fullscreen the current item
    public containers: Array<Container>; // Current set of visible containers
    public allD: Array<Container>; // All the containers we have loaded

    constructor(public _contentedService: ContentedService, public route: ActivatedRoute, public router: Router) {
    }

    public ngOnInit() {
        // Need to load content if the idx is greater than content loaded (n times potentially)
        this.route.paramMap.pipe().subscribe(
            (res: ParamMap) => {
                this.setPosition(
                    res.get('idx') ? parseInt(res.get('idx'), 10) : 0,
                    res.get('rowIdx') ? parseInt(res.get('rowIdx'), 10) : 0
                );
            },
            err => { console.error(err); }
        );
        this.setupEvtListener();
        this.calculateDimensions();
        this.loadDirs(); // Do this after the param map load potentially
    }

    // This will listen to nav events.
    public setupEvtListener() {
        GlobalNavEvents.navEvts.subscribe(evt => {
            switch(evt.action) {
                case NavTypes.NEXT_CONTAINER:
                    this.next();
                    break;
                case NavTypes.PREV_CONTAINER:
                    this.prev();
                    break;
                case NavTypes.LOAD_MORE:
                    this.loadMore();
                    break;
                case NavTypes.SELECT_MEDIA:
                    this.selectedMedia(evt.media, evt.cnt);
                    break;

                // Little awkward how this works
                case NavTypes.VIEW_FULLSCREEN:
                    this.viewFullscreen();
                    break;
                case NavTypes.HIDE_FULLSCREEN:
                    this.hideFullscreen();
                    break;
                case NavTypes.SAVE_MEDIA:
                    this.saveItem();
                    break;
                default:
                    break;
            }
            this.setCurrentItem();
        });
    }

    public saveItem() {
        console.log("We should save an item", this.getCurrentContainer());
        this._contentedService.download(this.getCurrentContainer(), this.rowIdx);
    }

    public loadMore() {
        let visible = this.getVisibleContainers();
        this.loadMoreInDir(visible[0]);
    }

    public viewFullscreen() {
        this.currentViewItem = this.getCurrentMedia();
        this.fullScreen = true;
    }

    public hideFullscreen() {
        this.fullScreen = false;
    }

    // Mostly for tests since testing full routing params is a god damn pain.
    public setPosition(idx: number, rowIdx: number) {
        this.idx = idx;
        this.rowIdx = rowIdx;
    }

    public loadDirs() {
        this.loading = true;
        this._contentedService.getPreview()
            .pipe(finalize(() => {this.loading = false; }))
            .subscribe(
                res => {
                    this.previewResults(res);
                },
                err => { console.error(err); }
            );
    }

    public loadMoreInDir(cnt: Container) {
        // This is being changed to just load more content up
        if (cnt.count < cnt.total && !this.loading) {
            this.loading = true;
            this._contentedService.loadMoreInDir(cnt)
                .pipe(finalize(() => {this.loading = false; }))
                .subscribe(
                    res => { this.cntResults(cnt, res); },
                    err => { console.error(err); }
                );
        }
    }

    public cntResults(cnt: Container, response) {
        console.log("Results loading, what is in the results?", response);
        cnt.addContents(cnt.buildImgs(response));
    }

    public reset() {
        this.idx = 0;
        this.allD = [];
        this.emptyMessage = null;
    }

    public getVisibleContainers() {
        if (this.allD) {
            let start = this.idx < this.allD.length ? this.idx : this.allD.length - 1;
            let end = start + this.maxVisible <= this.allD.length ? start + this.maxVisible : this.allD.length;

            let cnts = this.allD.slice(start, end);
            _.each(cnts, cnt => {
                this._contentedService.initialLoad(cnt);  // Only loads if cnt.loadState = LoadStates.NotLoaded
            });
            return cnts;
        }
        return [];
    }

    public setCurrentItem() {
        this.currentViewItem = this.getCurrentMedia();
        return this.currentViewItem;
    }

    public getCurrentContainer() {
        if (this.idx < this.allD.length && this.idx >= 0) {
            this.currentDir = this.allD[this.idx];
            return this.currentDir;
        }
        return null;
    }

    public updateRoute() {
        this.router.navigate([`/ui/browse/${this.idx}/${this.rowIdx}`]);
    }

    public nextMedia() {
        /*
        let cnt = this.getCurrentContainer();
        let items = cnt ? cnt.getContentList() : [];
        if (this.rowIdx < items.length) {
            this.rowIdx++;
            if (this.rowIdx === items.length && this.idx !== this.allD.length - 1) {
                this.next(true);
            }
        }
        this.setCurrentItem();
        */
    }

    public prevMedia() {
        /*
        if (this.rowIdx > 0) {
            this.rowIdx--;
        } else if (this.idx !== 0) {
            this.prev(true);
        }
        this.updateRoute();
        */
    }

    public next(selectFirst: boolean = true) {
        if (this.allD && this.idx + 1 < this.allD.length) {
            this.idx++;
        }
        if (selectFirst) {
            this.rowIdx = 0;
        }
        this.updateRoute();
    }

    public prev(selectLast: boolean = false) {
        if (this.idx > 0) {
            this.idx--;
        }
        if (selectLast) {
            let cnt = this.getCurrentContainer();
            let items = cnt ? cnt.getContentList() : [];
            this.rowIdx = items.length - 1;
        }
        this.updateRoute();
    }

    public imgLoaded(evt) {
        let img = evt.target;
        console.log("Img Loaded", img.naturalHeight, img.naturalWidth, img);
    }

    public getCurrentMedia() {
        let cnt = this.getCurrentContainer();
        if (cnt && !_.isEmpty(cnt.getContentList())) {
            let contentList = cnt.getContentList();
            if (this.rowIdx >= 0 && this.rowIdx < contentList.length) {
                return contentList[this.rowIdx];
            }
        }
    }

    // TODO: Being called abusively in the cntective rather than on page resize events
    @HostListener('window:resize', ['$event'])
    public calculateDimensions() {
        let width = !window['jasmine'] ? window.innerWidth : 800;
        let height = !window['jasmine'] ? window.innerHeight : 800;

        this.previewWidth = (width / 4) - 41;
        this.previewHeight = (height / this.maxVisible) - 41;
    }

    public previewResults(containers: Array<Container>) {
        console.log("Results returned from the preview results.", containers);
        this.allD = containers || [];
        if (_.isEmpty(containers)) {
            this.emptyMessage = "No Directories found, did you load the DB?";
        } else {
            this.loadView(this.idx, this.rowIdx);
        }
        return this.allD;
    }

    public fullLoadDir(cnt: Container) {
        this._contentedService.fullLoadDir(cnt).subscribe(
            (loadedDir: Container) => {
                console.log("Fully loaded up the container", loadedDir);
                this.setCurrentItem();
            },
            err => {console.error("Failed to load", err); }
        );
    }

    public loadView(idx, rowIdx) {
        let currDir = this.getCurrentContainer();
        if (rowIdx >= currDir.total) {
            rowIdx = 0;
        }
        this.idx = idx;
        this.rowIdx = rowIdx;

        if (rowIdx < currDir.count) {
            this.setCurrentItem();
        } else if (this.rowIdx < currDir.total) {
            this.fullLoadDir(currDir);
        }
    }

    public selectedMedia(media: Media, cnt: Container) {
        console.log("Click event, change currently selected indexes, container etc", media, cnt);
        let idx = _.findIndex(this.allD, {id: cnt ? cnt.id : -1});
        let rowIdx = cnt ? _.findIndex(cnt.contents, {id: media.id}) : -1;
        console.log("Found idx and row index: ", idx, rowIdx);
        if (idx >= 0 && rowIdx >= 0) {
            this.idx = idx;
            this.rowIdx = rowIdx;
            this.updateRoute();
        } else {
            console.error("Should not be able to click an item we cannot find.", cnt, media);
        }
    }
}
