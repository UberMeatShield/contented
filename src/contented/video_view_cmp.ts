import {forkJoin, Subscription} from 'rxjs';
import {finalize, debounceTime, map, distinctUntilChanged, catchError} from 'rxjs/operators';

import {
    OnInit,
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
import {ActivatedRoute, Router, ParamMap} from '@angular/router';
import {FormBuilder, NgForm, FormControl, FormGroup} from '@angular/forms';

import {PageEvent} from '@angular/material/paginator';
import {MatDialog, MatDialogConfig, MAT_DIALOG_DATA} from '@angular/material/dialog';
import * as _ from 'lodash';


@Component({
    selector: 'video-view-cmp',
    templateUrl: './video_view.ng.html'
})
export class VideoViewCmp implements OnInit{

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
        this.route.queryParams.pipe().subscribe(
            (res: ParamMap) => {
                let st = res['videoText'];
                let text = st !== undefined ? st : '';
                this.videoText.setValue(text);
                this.search(text); 
                this.setupFilterEvts();
            }
        );
        this.calculateDimensions();
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
        // pageIndex, pageSize
    }

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
        /*
        const dialogRef = this.dialog.open(
            SearchDialog,
            {
                data: mc,
                width: '90%',
                height: '100%',
                maxWidth: '100vw',
                maxHeight: '100vh',
            }
        );
        dialogRef.afterClosed().subscribe(result => {
            console.log("Closing the view", result);
        });
        */
    }

    imgLoaded(evt) {
        // Debugging / hooks but could also be a hook into a total loaded.
    }

    imgClicked(mc: Media) {
        console.log("Click the image", mc);
        this.fullView(mc);
    }
}