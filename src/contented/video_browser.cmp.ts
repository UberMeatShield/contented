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
import {Content} from './content';
import {Container} from './container';
import {Screen} from './screen';
import {GlobalNavEvents, NavTypes} from './nav_events';
import {ActivatedRoute, Router, ParamMap} from '@angular/router';
import {FormBuilder, NgForm, FormControl, FormGroup} from '@angular/forms';

import {PageEvent as PageEvent} from '@angular/material/paginator';
import {MatDialog as MatDialog, MatDialogConfig as MatDialogConfig, MAT_DIALOG_DATA} from '@angular/material/dialog';
import * as _ from 'lodash';


@Component({
    selector: 'video-browser-cmp',
    templateUrl: './video_browser.ng.html'
})
export class VideoBrowserCmp implements OnInit, OnDestroy {

    // Route needs to exist
    // Take in the search text route param
    // Debounce the search
    @ViewChild('searchForm', { static: true }) searchControl;
    throttleSearch: Subscription;
    videoText = new FormControl<string>("");
    options: FormGroup;
    fb: FormBuilder;

    public selectedContent: Content; // For keeping track of where we are in the page
    public selectedContainer: Container;  // For filtering
    public content: Array<Content>;
    public containers: Array<Container>;

    // TODO: Make this a saner calculation
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
    }

    ngOnDestroy() {
         if (this.sub) {
             this.sub.unsubscribe();
         }
     }

    public loadContainers() {
        this._contentedService.getContainers().subscribe(
            (cnts) => {
               this.containers = cnts.results || []; 
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
                     console.log("selectedContent", this.selectedContent, evt);
                     this.selectContent(this.selectedContent, this.selectedContainer);
                     break;
                 case NavTypes.LOAD_MORE:
                     // this.loadMore();
                     // It might not be TOO abusive to override this and make it page next?
                     break;
                 case NavTypes.SELECT_MEDIA:
                     this.selectContent(evt.content, evt.cnt);
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
        if (this.selectedContent && this.content) {
            let idx = _.findIndex(this.content, {id: this.selectedContent.id});
            if ((idx + 1) < this.content.length) {
                let m = this.content[idx+1];
                if (m.id != this.selectedContent.id) {
                    GlobalNavEvents.selectContent(m, new Container({id: m.container_id}));
                }       
            } else if ((this.offset + this.pageSize) < this.total) {
                this.search(this.videoText.value, (this.offset + this.pageSize), this.pageSize, this.getCntId());
            }
        }
    }

    public prev() {
        if (this.selectedContent && this.content) {
            let idx = _.findIndex(this.content, {id: this.selectedContent.id});
            if ((idx - 1) >= 0) {
                let m = this.content[idx-1];
                if (m.id != this.selectedContent.id) {
                    GlobalNavEvents.selectContent(m, new Container({id: m.container_id}));
                }
            } else if ((this.offset - this.pageSize) >= 0) {
                this.search(this.videoText.value, (this.offset - this.pageSize), this.pageSize, this.getCntId());
            }
        }
    }

    public selectContent(content: Content, container: Container) {
        this.selectedContent = content;
        console.log("Select content is executing.");
        _.delay(() => {
             let id = `view_content_${content.id}`;
             let el = document.getElementById(id)

             if (el) {
                 el.scrollIntoView(true);
                 window.scrollBy(0, -30);
             }
         }, 50);
    }


    public resetForm(setupFilterEvents: boolean = false) {
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
            // Prevent bubble on keypress
          )
          .subscribe(
              formData => {
                  console.log("Form data changing");
                  // If the text changes do we reset the search offset etc.
                  this.search(formData['videoText'] || '', 0, this.pageSize, this.getCntId());
              },
              error => {
                   console.error("failed to search, erro", error);
              }
          );
    }

    getValues() {
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

        this.selectedContent = null;
        this.content = [];
        this.loading = true;
        this._contentedService.searchContent(text, offset, limit, "video", cntId).pipe(
            finalize(() => this.loading = false)
        ).subscribe(
            (res) => {
                let content = _.map(res['content'], m => new Content(m));
                let total = res['total'] || 0;
                
                this.offset = offset;
                this.content = content;
                this.total = total;

                if (content && content.length > 0) {
                    let mc = content[0];
                    GlobalNavEvents.selectContent(mc, new Container({id: mc.container_id}));
                }
            }, err => {
                console.error("Failed to search for video content.", err);
            }
        );
    }

    // This will have to be updated to actually work
    public getVisibleSet() {
        return this.content;
    }

    public getContainer(cId: string) {
        return _.find(this.containers, {id: cId});
    }
}
