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
import {MediaContainer} from './directory';
import {ActivatedRoute, Router, ParamMap} from '@angular/router';
import {FormBuilder, NgForm, FormControl, FormGroup} from '@angular/forms';

import {PageEvent} from '@angular/material/paginator';
import {MatDialog, MatDialogConfig, MAT_DIALOG_DATA} from '@angular/material/dialog';
import * as _ from 'lodash';


@Component({
    selector: 'search-cmp',
    templateUrl: './search.ng.html'
})
export class SearchCmp implements OnInit{

    // Route needs to exist
    // Take in the search text route param
    // Debounce the search
    @ViewChild('searchForm', { static: true }) searchControl;
    throttleSearch: Subscription;
    searchText: FormControl;
    options: FormGroup;
    fb: FormBuilder;

    public media: Array<MediaContainer>;

    // TODO: Make this a saner calculation
    public previewWidth = 480;
    public previewHeight = 480;
    public maxVisible = 3; // How many results show horizontally
    public total = 0;
    public pageSize = 50;

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
                let st = res['searchText'];
                let text = st !== undefined ? st : '';
                console.log("Search text from url", text, res);
                this.searchText.setValue(text);
                if (text !== '') {
                    this.search(text); 
                }
                this.setupFilterEvts();
            }
        );
        this.calculateDimensions();
    }

    public resetForm(setupFilterEvents: boolean = false) {
        this.searchText = new FormControl('');
        this.options = this.fb.group({
            searchText: this.searchText,
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
                  this.search(formData['searchText'] || '');
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
        console.log("Event", evt, this.searchText.value);
        let offset = evt.pageIndex * evt.pageSize;
        let limit = evt.pageSize;
        this.search(this.searchText.value, offset, limit);
        // pageIndex, pageSize
    }

    public search(text: string, offset: number = 0, limit: number = 50) {
        console.log("Get the information from the input and search on it", text); 
        // TODO: Wrap the media into a fake directory
        this.media = [];
        this._contentedService.searchMedia(text, offset, limit).subscribe(
            (res) => {
                let media = _.map((res['media'] || []), m => new MediaContainer(m));
                let total = res['total'] || 0;
                
                console.log("Search results", media, total);
                this.media = media;
                this.total = total;
            }, err => {
                console.error("Failed to search", err);
            }
        );
    }

    public getVisibleSet() {
        return this.media;
    }

    // TODO: Being called abusively in the directive rather than on page resize events
    @HostListener('window:resize', ['$event'])
    public calculateDimensions() {
        let width = !window['jasmine'] ? window.innerWidth : 800;
        let height = !window['jasmine'] ? window.innerHeight : 800;

        this.previewWidth = (width / 4) - 41;
        this.previewHeight = (height / this.maxVisible) - 41;
    }

    public fullView(mc: MediaContainer) {
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
    }

    imgLoaded(evt) {
        // Debugging / hooks but could also be a hook into a total loaded.
    }

    imgClicked(mc: MediaContainer) {
        console.log("Click the image", mc);
        this.fullView(mc);
    }
}

// This just doesn't seem like a great approach :(
@Component({
    selector: 'search-dialog',
    templateUrl: 'search_dialog.ng.html'
})
export class SearchDialog implements AfterViewInit {

    public mediaContainer: MediaContainer;

    public forceHeight: number;
    public forceWidth: number;
    public sizeCalculated: boolean = false;

    @ViewChild('SearchContent', { static: true }) searchContent;

    constructor(@Inject(MAT_DIALOG_DATA) public mc: MediaContainer, public _service: ContentedService) {
        // console.log("Mass taker opened with items:", items);
        this.mediaContainer = mc;
    }

    ngAfterViewInit() {
        console.log("Search content is:", this.searchContent);

        setTimeout(() => {
            let el = this.searchContent.nativeElement;
            if (el) {
                console.log("Element", el, el.offsetWidth, el.offsetHeight);
                this.forceHeight = el.offsetHeight;
                this.forceWidth = el.offsetWidth;
            }
            this.sizeCalculated = true;
        }, 100);
    }
}