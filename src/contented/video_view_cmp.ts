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

    public media: Array<Media>;

    // TODO: Make this a saner calculation
    public previewWidth = 480;
    public previewHeight = 480;
    public screenWidth = 960;
    public maxVisible = 3; // How many results show horizontally
    public total = 0;
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
                this.videoText.setValue(text);
                this.search(text); 
                this.setupFilterEvts();
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
                 case NavTypes.LOAD_MORE:
                     // this.loadMore();
                     // It might not be TOO abusive to override this and make it page next?
                     break;
                 case NavTypes.SELECT_MEDIA:
                     // this.selectedMedia(evt.media, evt.cnt);
                     break;
                 case NavTypes.SELECT_CONTAINER:
                     // this.selectContainer(evt.cnt);
                     break;
                 default:
                     break;
             }
         });
     }

    public next() {
        console.log("Next");
        // It should have a jump to scroll location for the currently selected item
    }

    public prev() {
        console.log("Previous");
    }

    public selectMedia(media: Media) {
        console.log("Selected media", media);
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
                  this.search(formData['videoText'] || '');
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
        this.search(this.videoText.value, offset, limit);
    }


    // TODO: Add in optional filter params like the container (filter by container in search?)
    public search(text: string, offset: number = 0, limit: number = 50) {
        console.log("Get the information from the input and search on it", text); 
        // TODO: Wrap the media into a fake container
        this.media = [];
        this.loading = true;
        this._contentedService.searchMedia(text, offset, limit, "video").pipe(
            finalize(() => this.loading = false)
        ).subscribe(
            (res) => {
                let media = _.map(res['media'], m => new Media(m));
                let total = res['total'] || 0;
                
                this.media = media;
                this.total = total;
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
                this.forceHeight = el.offsetHeight;
                this.forceWidth = el.offsetWidth;
            }
            this.sizeCalculated = true;
        }, 100);
    }
}
