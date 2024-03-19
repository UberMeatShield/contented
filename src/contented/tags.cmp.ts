import {Subscription} from 'rxjs';
import {debounceTime, distinctUntilChanged, } from 'rxjs/operators';

import {
    OnInit,
    Component,
    Input,
    ViewChild,
} from '@angular/core';
import { Tag } from './content';
import {ContentedService} from './contented_service';
import {FormBuilder, FormControl, FormGroup} from '@angular/forms';

import {PageEvent} from '@angular/material/paginator';
import * as _ from 'lodash';
import { GlobalBroadcast } from './global_message';

@Component({
    selector: 'tags-cmp',
    templateUrl: './tags.ng.html'
})
export class TagsCmp implements OnInit{

    // Route needs to exist
    // Take in the search text route param
    // Debounce the search
    @ViewChild('searchForm', { static: true }) searchControl;

    @Input() tags: Array<Tag>;
    @Input() loadTags = false;

    matchedTags: Array<Tag>;
    throttleSearch: Subscription;
    searchTags = new FormControl<string>("");
    options: FormGroup;
    fb: FormBuilder;


    public loading: boolean = false;
    public pageSize: number = 1000;
    public total: number = 0;

    constructor(
        public _contentedService: ContentedService,
        fb: FormBuilder,
    ) {
        this.fb = fb;
    }

    public ngOnInit() {
        if (this.loadTags) {
            this.search('');
        }
        this.resetForm(true);
    }

    // TODO: Get the current input token
    // TODO: Suggest input tokens
    // TODO: Provide the ability to select tokens and also remove a token from VSCode

    public search(searchText: string) {
        this._contentedService.getTags().subscribe({
            next: (res: {results: Array<Tag>, total: number}) => {
                this.tags = res.results;
            },
            error: err => {
                GlobalBroadcast.error('Failed to load tags', err);
            }
        });
    }

    // Change the event to provide both the value and the parsed tags
    public changedTags(evt) {
        console.log("Changed Tags", evt);
        //this.search(evt)
    }

    public resetForm(setupFilterEvents: boolean = false) {
        this.options = this.fb.group({
            searchTags: this.searchTags,
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
          .subscribe({
              next: formData => {
                const searchTag = formData.searchTags;
                console.log("It should setup filter events", searchTag);
                if (searchTag) {
                    const matcher = new RegExp(searchTag, 'ig');
                    this.matchedTags = _.filter(this.tags, t => matcher.test(t.id));
                } else {
                    this.matchedTags = this.tags;
                }
              },
              error: err => {
                GlobalBroadcast.error('Failed to load tags', err);
              }
          });
    }

    public getValues() {
        return this.options.value;
    }

    pageEvt(evt: PageEvent) {
        console.log("Event", evt, this.searchTags.value);
        let offset = evt.pageIndex * evt.pageSize;
        let limit = evt.pageSize;
    }
}