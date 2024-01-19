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
import {Content} from './content';
import {ActivatedRoute, Router, ParamMap} from '@angular/router';
import {FormBuilder, NgForm, FormControl, FormGroup} from '@angular/forms';

import {PageEvent} from '@angular/material/paginator';
import {MatDialog, MatDialogConfig, MAT_DIALOG_DATA} from '@angular/material/dialog';
import * as _ from 'lodash';


@Component({
    selector: 'earcearch-cmp',
    templateUrl: './search.ng.html'
})
export class SearchCmp implements OnInit{

    // Route needs to exist
    // Take in the search text route param
    // Debounce the search
    @ViewChild('videoForm', { static: true }) searchControl;
    throttleSearch: Subscription;
    searchText = new FormControl<string>("");
    options: FormGroup;
    fb: FormBuilder;

    public content: Array<Content>;

    // TODO: Make this a saner calculation
    public previewWidth = 480;
    public previewHeight = 480;
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
                let st = res['searchText'];
                let text = st !== undefined ? st : '';
                console.log("Search text from url", text, res);
                this.searchText.setValue(text);
                this.search(text); 
                this.setupFilterEvts();
            }
        );
        this.calculateDimensions();
    }

    public resetForm(setupFilterEvents: boolean = false) {
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
    }

    public search(text: string, offset: number = 0, limit: number = 50) {
        console.log("Get the information from the input and search on it", text); 
        // TODO: Wrap the content into a fake container
        this.content = [];
        this.loading = true;
        this._contentedService.searchContent(text, offset, limit).pipe(
            finalize(() => this.loading = false)
        ).subscribe(
            (res) => {
                let content = _.map((res.results || []), m => new Content(m));
                let total = res['total'] || 0;
                // console.log("Search results", content, total);
                this.content = content;
                this.total = total;
            }, err => {
                console.error("Failed to search", err);
            }
        );
    }

    public getVisibleSet() {
        return this.content;
    }

    // TODO: Being called abusively in the cntective rather than on page resize events
    @HostListener('window:resize', ['$event'])
    public calculateDimensions() {
        let width = !window['jasmine'] ? window.innerWidth : 800;
        let height = !window['jasmine'] ? window.innerHeight : 800;

        this.previewWidth = (width / 4) - 41;
        this.previewHeight = (height / this.maxVisible) - 41;
    }

    public fullView(mc: Content) {
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

    contentClicked(mc: Content) {
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

    public contentContainer: Content;

    public forceHeight: number;
    public forceWidth: number;
    public sizeCalculated: boolean = false;

    @ViewChild('SearchContent', { static: true }) searchContent;

    constructor(@Inject(MAT_DIALOG_DATA) public mc: Content, public _service: ContentedService) {
        // console.log("Mass taker opened with items:", items);
        this.contentContainer = mc;
    }

    ngAfterViewInit() {
        // TODO: Sizing content is a little off and the toolbars are visible based on dialog size
        setTimeout(() => {
            let el = this.searchContent.nativeElement;
            if (el) {
                console.log("Element", el, el.offsetWidth, el.offsetHeight);
                this.forceHeight = el.offsetHeight - 40;
                this.forceWidth = el.offsetWidth - 40;
            }
            this.sizeCalculated = true;
        }, 100);
    }
}
