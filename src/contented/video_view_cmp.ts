import {forkJoin, Subscription} from 'rxjs';
import {finalize, debounceTime, map, distinctUntilChanged, catchError} from 'rxjs/operators';

import {
    OnInit,
    OnDestroy,
    AfterViewInit,
    Component,
    EventEmitter,
    Input,
    Output,
    HostListener,
    ViewChild,
    Inject
} from '@angular/core';
import {ContentedService} from './contented_service';
import {Media} from './media';
import {Container} from './container';
import {Screen} from './screen';
import {GlobalNavEvents, NavTypes} from './nav_events';
import {ActivatedRoute, Router, ParamMap} from '@angular/router';
import {FormBuilder, NgForm, FormControl, FormGroup} from '@angular/forms';

import {PageEvent} from '@angular/material/paginator';
import {MatDialog, MatDialogConfig, MAT_DIALOG_DATA} from '@angular/material/dialog';
import * as _ from 'lodash';


@Component({
    selector: 'video-view-cmp',
    templateUrl: './video_view.ng.html'
})
export class VideoViewCmp implements OnInit, OnDestroy {

    // Route needs to exist
    // Take in the search text route param
    // Debounce the search
    @ViewChild('searchForm', { static: true }) searchControl;
    throttleSearch: Subscription;
    videoText: FormControl;
    options: FormGroup;
    fb: FormBuilder;

    public selectedMedia: Media; // For keeping track of where we are in the page
    public selectedContainer: Container;  // For filtering
    public media: Array<Media>;
    public containers: Array<Container>;

    // TODO: Make this a saner calculation
    public previewWidth = 480;
    public previewHeight = 480;
    public screenWidth = 960;
    public maxVisible = 3; // How many results show vertically
    public total = 0;
    public offset = 0; // Tracking where we are in the position
    public pageSize = 50;
    public loading: boolean = false;
    public sub: Subscription;  // Listening for GlobalNavEvents

    constructor(
        public _contentedService: ContentedService,
        public route: ActivatedRoute,
        public router: Router,
        public dialog: MatDialog,
        fb: FormBuilder,
    ) {
        this.fb = fb;
    }

    public ngOnInit() {
        this.resetForm();

        // This should also preserve the current page we have selected and restore it.
        this.route.queryParams.pipe().subscribe(
            (res: ParamMap) => {
                let st = res['videoText'];
                let text = st !== undefined ? st : '';

                // Add in a param for container_id ?

                this.videoText.setValue(text);
                this.search(text, this.offset, this.pageSize, this.getCntId()); 
                this.setupFilterEvts();
                this.loadContainers();
            }
        );
        this.setupEvtListener();
        this.calculateDimensions();
    }

    ngOnDestroy() {
         if (this.sub) {
             this.sub.unsubscribe();
         }
     }

    public loadContainers() {
        this._contentedService.getContainers().subscribe(
            (cnts: Array<Container>) => {
               this.containers = cnts; 
            }
        );
    }
 
     // This will listen to nav events.
     public setupEvtListener() {
         this.sub = GlobalNavEvents.navEvts.subscribe(evt => {
             switch(evt.action) {
                 case NavTypes.NEXT_MEDIA:
                     this.next();
                     break;
                 case NavTypes.PREV_MEDIA:
                     this.prev();
                     break;
                 case NavTypes.HIDE_FULLSCREEN:
                     // Scroll back into view
                     this.selectMedia(this.selectedMedia, this.selectedContainer);
                     break;
                 case NavTypes.LOAD_MORE:
                     // this.loadMore();
                     // It might not be TOO abusive to override this and make it page next?
                     break;
                 case NavTypes.SELECT_MEDIA:
                     this.selectMedia(evt.media, evt.cnt);
                     break;
                 case NavTypes.SELECT_CONTAINER:
                     this.selectContainer(evt.cnt);
                     break;
                 default:
                     break;
             }
         });
     }

    public selectContainer(cnt: Container) {
        let offset = this.offset;
        if (_.get(cnt, 'id') != _.get(this.selectedContainer, 'id')) {
            this.offset = 0;
        }
        this.selectedContainer = cnt;
        this.search(this.videoText.value, this.offset, this.pageSize, this.getCntId());
    }

    public next() {
        // It should have a jump to scroll location for the currently selected item
        if (this.selectedMedia && this.media) {
            let idx = _.findIndex(this.media, {id: this.selectedMedia.id});
            if ((idx + 1) < this.media.length) {
                let m = this.media[idx+1];
                if (m.id != this.selectedMedia.id) {
                    GlobalNavEvents.selectMedia(m, new Container({id: m.container_id}));
                }       
            } else if ((this.offset + this.pageSize) < this.total) {
                this.search(this.videoText.value, (this.offset + this.pageSize), this.pageSize, this.getCntId());
            }
        }
    }

    public prev() {
        if (this.selectedMedia && this.media) {
            let idx = _.findIndex(this.media, {id: this.selectedMedia.id});
            if ((idx - 1) >= 0) {
                let m = this.media[idx-1];
                if (m.id != this.selectedMedia.id) {
                    GlobalNavEvents.selectMedia(m, new Container({id: m.container_id}));
                }
            } else if ((this.offset - this.pageSize) >= 0) {
                this.search(this.videoText.value, (this.offset - this.pageSize), this.pageSize, this.getCntId());
            }
        }
    }

    public selectMedia(media: Media, container: Container) {
        this.selectedMedia = media;
        _.delay(() => {
             let id = `view_media_${media.id}`;
             let el = document.getElementById(id)

             if (el) {
                 el.scrollIntoView(true);
                 window.scrollBy(0, -30);
             }
         }, 20);

    }


    public resetForm(setupFilterEvents: boolean = false) {
        this.videoText = new FormControl('');
        this.options = this.fb.group({
            videoText: this.videoText,
        });
        if (setupFilterEvents) {
            this.setupFilterEvts();
        }
    }

    public setupFilterEvts() {
        // Kicks off a search
        if (this.throttleSearch) {
            this.throttleSearch.unsubscribe();
        }
        this.throttleSearch = this.options.valueChanges
          .pipe(
            debounceTime(500),
            distinctUntilChanged()
          )
          .subscribe(
              formData => {
                  // If the text changes do we reset the search offset etc.
                  this.search(formData['videoText'] || '', 0, this.pageSize, this.getCntId());
              },
              error => {
                   console.error("failed to search, erro", error);
              }
          );
    }

    public getValues() {
        return this.options.value;
    }

    pageEvt(evt: PageEvent) {
        console.log("Event", evt, this.videoText.value);
        let offset = evt.pageIndex * evt.pageSize;
        let limit = evt.pageSize;
        this.search(this.videoText.value, offset, limit, this.getCntId());
    }

    public getCntId() {
        return !!this.selectedContainer ? this.selectedContainer.id : null;
    }


    // TODO: Add in optional filter params like the container (filter by container in search?)
    public search(text: string, offset: number = 0, limit: number = 50, cntId: string = null) {
        console.log("Get the information from the input and search on it", text, offset, limit, cntId); 

        this.selectedMedia = null;
        this.media = [];
        this.loading = true;
        this._contentedService.searchMedia(text, offset, limit, "video", cntId).pipe(
            finalize(() => this.loading = false)
        ).subscribe(
            (res) => {
                let media = _.map(res['media'], m => new Media(m));
                let total = res['total'] || 0;
                
                this.offset = offset;
                this.media = media;
                this.total = total;

                if (media && media.length > 0) {
                    GlobalNavEvents.selectMedia(media[0], new Container({id: media.container_id}));
                }
            }, err => {
                console.error("Failed to search for video media.", err);
            }
        );
    }

    // This will have to be updated to actually work
    public getVisibleSet() {
        return this.media;
    }

    // TODO: Being called abusively in the cntective rather than on page resize events
    @HostListener('window:resize', ['$event'])
    public calculateDimensions() {
        let width = !window['jasmine'] ? window.innerWidth : 800;
        let height = !window['jasmine'] ? window.innerHeight : 800;

        this.previewWidth = width / 5;
        this.previewHeight = (height / this.maxVisible) - 41;

        // screenHeight is just calculated on the component previewHeight * 2
        this.screenWidth = width - this.previewWidth - 200;  // Fudge factor
    }

    public fullView(mc: Media) {
        console.log("Full view", mc);
        GlobalNavEvents.viewFullScreen(mc);
    }

    public screenEvt(evt) {
        console.log("Screen Evt", evt);
        const dialogRef = this.dialog.open(
            ScreenDialog,
            {
                data: {screen: evt.screen, screens: evt.screens},
                width: '90%',
                height: '100%',
                maxWidth: '100vw',
                maxHeight: '100vh',
            }
        );
        dialogRef.afterClosed().subscribe(result => {
            console.log("Closing the view", result);
        });
    }

    imgLoaded(evt) {
        // Debugging / hooks but could also be a hook into a total loaded.
    }

    imgClicked(mc: Media) {
        console.log("Click the image", mc);
        this.fullView(mc);
    }
}

// This just doesn't seem like a great approach :(
@Component({
    selector: 'screen-dialog',
    templateUrl: 'screen_dialog.ng.html'
})
export class ScreenDialog implements AfterViewInit {

    public screen: Screen;
    public screens: Array<Screen>

    public forceHeight: number;
    public forceWidth: number;
    public sizeCalculated: boolean = false;
    @ViewChild('ScreensContent', { static: true }) screenContent;
    
    constructor(
        @Inject(MAT_DIALOG_DATA) public data,
        public _service: ContentedService) {

        this.screen = data.screen;
        this.screens = data.screens;
    }

    ngAfterViewInit() {
        console.log("Search content is:", this.screenContent);
        setTimeout(() => {
            let el = this.screenContent.nativeElement;
            if (el) {
                console.log("Element", el, el.offsetWidth, el.offsetHeight);
                this.forceHeight = el.offsetHeight - 100;
                this.forceWidth = el.offsetWidth - 100;
            }
            this.sizeCalculated = true;
        }, 100);

        this.idx();
    }

    idx() {
        if (this.screens && this.screen) {
            return _.findIndex(this.screens, {id: this.screen.id});
        }
        return -1;
    }

    next() {
        let i = this.idx();
        if (i < this.screens.length - 1) {
            this.screen = this.screens[i + 1];
        }
    }

    prev() {
        let i = this.idx();
        if (i - 1 >=  0) {
            this.screen = this.screens[i - 1];
        }
    }
}
