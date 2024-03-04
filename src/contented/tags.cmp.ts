import {Subscription} from 'rxjs';
import {debounceTime, distinctUntilChanged, } from 'rxjs/operators';

import {
    OnInit,
    Component,
    Output,
    ViewChild,
} from '@angular/core';
import {ContentedService} from './contented_service';
import {FormBuilder, FormControl, FormGroup} from '@angular/forms';

import {PageEvent} from '@angular/material/paginator';
import * as _ from 'lodash';

@Component({
    selector: 'tags-cmp',
    templateUrl: './tags.ng.html'
})
export class TagsCmp implements OnInit{

    // Route needs to exist
    // Take in the search text route param
    // Debounce the search
    @ViewChild('searchForm', { static: true }) searchControl;
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
        this.resetForm();
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
          .subscribe(
              formData => {
                console.log("It should setup filter events");
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
        console.log("Event", evt, this.searchTags.value);
        let offset = evt.pageIndex * evt.pageSize;
        let limit = evt.pageSize;
    }
}