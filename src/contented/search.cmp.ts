import { Subscription } from 'rxjs';
import { finalize, debounceTime, distinctUntilChanged } from 'rxjs/operators';

import { Input, OnInit, AfterViewInit, Component, HostListener, ViewChild, Inject } from '@angular/core';
import { ContentedService, ContentSearchSchema } from './contented_service';
import { Content, Tag, VSCodeChange } from './content';
import { ActivatedRoute, Router, ParamMap } from '@angular/router';
import { FormBuilder, FormGroup, FormControl } from '@angular/forms';

import { PageEvent } from '@angular/material/paginator';
import { MatDialog, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { GlobalBroadcast } from './global_message';
import * as _ from 'lodash';

@Component({
  selector: 'search-cmp',
  templateUrl: './search.ng.html',
})
export class SearchCmp implements OnInit {
  // Route needs to exist
  // Take in the search text route param
  // Debounce the search
  @ViewChild('videoForm', { static: true }) searchControl;
  @Input() tags: Array<Tag>;

  throttleSearch: Subscription;
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

  public searchText: string; // Initial searchText value if passed in the url
  public searchType = new FormControl('text');
  public currentTextChange: VSCodeChange = { value: '', tags: [] };
  public changedSearch: (evt: VSCodeChange) => void;

  constructor(
    public _contentedService: ContentedService,
    public route: ActivatedRoute,
    public router: Router,
    public dialog: MatDialog,
    fb: FormBuilder
  ) {
    this.fb = fb;
    this.options = fb.group({
      searchType: this.searchType,
    });
  }

  public ngOnInit() {
    // We don't want to call search ever keypress and changeSearch is being called
    // by an event emitter with a different debounce & distinct timing.
    this.changedSearch = _.debounce((evt: VSCodeChange) => {
      // Do not change this.searchText it will re-assign the VS-Code editor in a
      // bad way and muck with the cursor.
      this.search(evt.value, 0, 50, evt.tags);
      this.currentTextChange = evt;
    }, 250);

    this.route.queryParams.pipe().subscribe({
      next: (res: ParamMap) => {
        console.log('Query Params set', res);
        // Note you do NOT want searchText to be updated by changes
        // in this component except possibly a 'clear'
        this.searchText = res['searchText'] || '';
      },
    });
    this.calculateDimensions();
    this.setupFilterEvts();
  }

  /*
   * Should reset the pagination utils?
   */
  public changeSearch(evt: VSCodeChange) {
    /* DO NOT re-assign searchText or it will reassign the VSCode variable
        this.searchText = evt.value;
        this.searchTags = evt.tags;
        */
    this.changedSearch(evt);
  }

  // TODO: Need to throttle the changes to the changeSearch from VSCode and
  // remove some of this form based data (or create a hidden form with other settings)
  public setupFilterEvts() {
    if (this.throttleSearch) {
      this.throttleSearch.unsubscribe();
    }

    // This will need to be implemented once there are more controls in place.
    this.throttleSearch = this.options.valueChanges.pipe(debounceTime(250), distinctUntilChanged()).subscribe({
      next: (formData: FormData) => {
        // Eventually the form probably will have some data
        const evt = this.currentTextChange;
        this.search(evt?.value, 0, 50, evt?.tags);
      },
      error: err => {
        GlobalBroadcast.error('Failed to search', err);
      },
    });
  }

  pageEvt(evt: PageEvent) {
    console.log('Event', evt, this.searchText);
    let offset = evt.pageIndex * evt.pageSize;
    let limit = evt.pageSize;
    this.search(this.currentTextChange.value, offset, limit, this.currentTextChange.tags);
  }

  public search(text: string, offset: number = 0, limit: number = 50, tags: Array<string> = []) {
    console.log('Get the information from the input and search on it', text);
    // TODO: Wrap the content into a fake container
    this.content = [];
    this.loading = true;

    // TODO: Make this a bit less sketchy after I work on the actual data tagging.
    const searchType = this.options.get('searchType').value;
    if (searchType === 'tags') {
      text = '';
    } else {
      tags = [];
    }
    // TODO: Make the tags optional
    const cs = ContentSearchSchema.parse({ text, offset, limit, tags });
    this._contentedService
      .searchContent(cs)
      .pipe(finalize(() => (this.loading = false)))
      .subscribe({
        next: res => {
          let content = _.map(res.results || [], m => new Content(m));
          let total = res['total'] || 0;
          // console.log("Search results", content, total);
          this.content = content;
          this.total = total;
        },
        error: err => {
          console.error('Failed to search', err);
        },
      });
  }

  public getVisibleSet() {
    return this.content || [];
  }

  // TODO: Being called abusively in the content rather than on page resize events
  @HostListener('window:resize', ['$event'])
  public calculateDimensions() {
    let width = !window['jasmine'] ? window.innerWidth : 800;
    let height = !window['jasmine'] ? window.innerHeight : 800;

    this.previewWidth = width / 4 - 41;
    this.previewHeight = height / this.maxVisible - 41;
  }

  /**
   * Don't really love the dialog, might want to swap this or need to cleanup the dialog
   * @param mc
   */
  public fullView(mc: Content) {
    const dialogRef = this.dialog.open(SearchDialog, {
      data: mc,
      width: '90%',
      height: '100%',
      maxWidth: '100vw',
      maxHeight: '100vh',
    });
    dialogRef.afterClosed().subscribe(result => {
      console.log('Closing the view', result);
    });
  }

  imgLoaded(evt) {
    // Debugging / hooks but could also be a hook into a total loaded.
  }

  contentClicked(mc: Content) {
    console.log('Click the image', mc);
    this.fullView(mc);
  }
}

// This just doesn't seem like a great approach :(
@Component({
  selector: 'search-dialog',
  templateUrl: 'search_dialog.ng.html',
})
export class SearchDialog implements AfterViewInit {
  public contentContainer: Content;

  public forceHeight: number;
  public forceWidth: number;
  public sizeCalculated: boolean = false;

  @ViewChild('SearchContent', { static: true }) searchContent;

  constructor(
    @Inject(MAT_DIALOG_DATA) public mc: Content,
    public _service: ContentedService
  ) {
    // console.log("Mass taker opened with items:", items);
    this.contentContainer = mc;
  }

  ngAfterViewInit() {
    // TODO: Sizing content is a little off and the toolbars are visible based on dialog size
    setTimeout(() => {
      let el = this.searchContent.nativeElement;
      if (el) {
        console.log('Element', el, el.offsetWidth, el.offsetHeight);
        this.forceHeight = el.offsetHeight - 40;
        this.forceWidth = el.offsetWidth - 40;
      }
      this.sizeCalculated = true;
    }, 100);
  }
}
