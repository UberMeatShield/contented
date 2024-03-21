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


    public loading: boolean = false;
    public pageSize: number = 1000;
    public total: number = 0;

    constructor(
        public _contentedService: ContentedService,
    ) {
    }

    public ngOnInit() {
        if (this.loadTags) {
            this.search('');
        }
    }

    // TODO: Get the current input token
    // TODO: Suggest input tokens
    // TODO: Provide the ability to select tokens and also remove a token from VSCode
    public search(searchText: string) {
        this._contentedService.getTags().subscribe({
            next: (res: any ) => {
                this.tags = res.results;
            },
            error: (err) => {
                GlobalBroadcast.error('Failed to load tags in tagging component', err.message);
            }
        });
    }

    // Change the event to provide both the value and the parsed tags
    public changedTags(evt) {
        console.log("Changed Tags", evt);
        //this.search(evt)
    }
}