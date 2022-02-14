import {Subscription} from 'rxjs';
import {OnInit, OnDestroy, Component, EventEmitter, Input, Output, HostListener} from '@angular/core';
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
export class ContentedCmp implements OnInit, OnDestroy {

    @Input() maxVisible: number = 2; // How many of the loaded containers should we be viewing
    @Input() rowIdx: number = 0; // Which row (media item) are we on
    @Input() idx: number = 0; // Which item within the container are we viewing

    public loading: boolean = false;
    public emptyMessage = null;
    public previewWidth: number = 200; // Based on current client page sizes, scale the preview images natually
    public previewHeight: number = 200; // height for the previews ^

    // TODO: Remove this listener
    public fullScreen: boolean = false; // Should we view fullscreen the current item
    public containers: Array<Container>; // Current set of visible containers
    public allD: Array<Container>; // All the containers we have loaded
    public sub: Subscription;

    constructor(public _contentedService: ContentedService, public route: ActivatedRoute, public router: Router) {
    }

    public ngOnInit() {
        // Need to load content if the idx is greater than content loaded (n times potentially)
        this.route.paramMap.pipe().subscribe(
            (res: ParamMap) => {
                this.setPosition(
                    res.get('idx') ? parseInt(res.get('idx'), 10) : this.idx,
                    res.get('rowIdx') ? parseInt(res.get('rowIdx'), 10) : 0,
                );
            },
            err => { console.error(err); }
        );
        this.setupEvtListener();
        this.calculateDimensions();
        this.loadDirs(); // Do this after the param map load potentially
    }

    ngOnDestroy() {
        if (this.sub) {
            this.sub.unsubscribe();
        }
    }

    // This will listen to nav events.
    public setupEvtListener() {
        this.sub = GlobalNavEvents.navEvts.subscribe(evt => {
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
                default:
                    break;
            }
        });
    }

    public loadMore() {
        let visible = this.getVisibleContainers();
        this.loadMoreInDir(visible[0]);
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
            // Only loads if cnt.loadState = LoadStates.NotLoaded
            let currCnt = this.getCurrentContainer();
            let cnts = this.allD.slice(start, end);
            _.each(cnts, (cnt, idx) => {
                let obs = this._contentedService.initialLoad(cnt); 
                if (obs) { 
                    obs.subscribe(
                        media => {
                            if (cnt == currCnt) {
                                GlobalNavEvents.selectMedia(cnt.getMedia(), cnt);
                            }
                        }, err => console.error
                    );
                }
            });
            return cnts;
        }
        return [];
    }

    public getCurrentContainer() {
        if (this.idx < this.allD.length && this.idx >= 0) {
            return this.allD[this.idx];
        }
        return null;
    }

    public updateRoute() {
        let cnt = this.allD[this.idx];
        this.router.navigate([`/ui/browse/${this.idx}/${cnt.rowIdx}`]);
    }

    public next(selectFirst: boolean = true) {
        if (this.allD && this.idx + 1 < this.allD.length) {
            this.idx++;
            let cnt = this.getCurrentContainer();
            GlobalNavEvents.selectMedia(cnt.getMedia(), cnt);
            this.updateRoute();
        }
    }

    public prev(selectLast: boolean = false) {
        if (this.idx > 0) {
            this.idx--;
            let cnt = this.getCurrentContainer();

            console.log("PREVIOUS", cnt, cnt.getMedia());
            GlobalNavEvents.selectMedia(cnt.getMedia(), cnt);
            this.updateRoute();
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
            // Maybe just read the current param from the route 
            this.loadView(this.idx, this.rowIdx, true);
        }
        return this.allD;
    }

    public fullLoadDir(cnt: Container) {
        console.log("Full load directory");
        this._contentedService.fullLoadDir(cnt).subscribe(
            (loadedCnt: Container) => {
                console.log("Fully loaded up the container", loadedCnt);
                GlobalNavEvents.selectMedia(loadedCnt.getMedia(), loadedCnt);
            },
            err => {console.error("Failed to load", err); }
        );
    }

    public loadView(idx: number, rowIdx: number, triggerSelect: boolean = false) {
        let currDir = this.getCurrentContainer();
        if (rowIdx >= currDir.total) {
            rowIdx = 0;
        }
        this.idx = idx;
        currDir.rowIdx = rowIdx;

        // This handles the case where we need to fully load a container to reach the row
        if (rowIdx > currDir.count) {
            console.log("This shoudl load around currDir", currDir);
            this.fullLoadDir(currDir);
        } else if (triggerSelect) {
            let cnt = this.getCurrentContainer();
            GlobalNavEvents.selectMedia(cnt.getMedia(), cnt);
        }
    }

    // Could probably move this into a saner location
    public selectedMedia(media: Media, cnt: Container) {
        //console.log("Click event, change currently selected indexes, container etc", media, cnt);
        let idx = _.findIndex(this.allD, {id: cnt ? cnt.id : -1});
        if (idx >= 0) {
            this.idx = idx;
            this.rowIdx = cnt.rowIdx;
            this.updateRoute();
        } else {
            console.error("Should not be able to click an item we cannot find.", cnt, media);
        }
    }
}
