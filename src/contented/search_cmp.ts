import {forkJoin, Subscription} from 'rxjs';
import {finalize, debounceTime, map, distinctUntilChanged, catchError} from 'rxjs/operators';

import {OnInit, Component, EventEmitter, Input, Output, HostListener, ViewChild} from '@angular/core';
import {ContentedService} from './contented_service';
import {MediaContainer} from './directory';

import {ActivatedRoute, Router, ParamMap} from '@angular/router';
import {FormBuilder, NgForm, FormControl, FormGroup} from '@angular/forms';


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

    constructor(
        public _contentedService: ContentedService,
        public route: ActivatedRoute,
        public router: Router,
        fb: FormBuilder,
    ) {
        this.fb = fb;
    }

    public ngOnInit() {
        this.resetForm();
        this.route.paramMap.pipe().subscribe(
            (res: ParamMap) => {
                let text = res.get("text");
                console.log("Search text from url", text);
                this.searchText.setValue(text);
                this.setupFilterEvts();
            }
        );

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

    public search(text) {
       console.log("Get the information from the input and search on it", text); 
    }
}
