import {OnInit, Component, EventEmitter, Input, Output, HostListener} from '@angular/core';
import {ContentedService} from './contented_service';

import {Directory} from './directory';
import * as _ from 'lodash';

@Component({
    selector: 'directory-cmp',
    templateUrl: 'directory.ng.html'
})
export class DirectoryCmp implements OnInit {

    @Input() dir: Directory;
    @Input() previewWidth: number;
    @Input() previewHeight: number;

    @Input() currentViewItem: string;
    @Input() maxRendered: number = 8; // Default setting for how many should be visible at any given time
    @Input() maxPrevItems: number = 2; // When scrolling through a dir, how many previous items should be visible

    @Output() clickedItem: EventEmitter<any> = new EventEmitter<any>();

    // @Output clickEvt: EventEmitter<any>;
    public visibleSet: Array<string>; // The currently visible set of items from in the directory

    constructor() {

    }

    public ngOnInit() {
        console.log("Directory Component loading up");
    }


    public getVisibleSet(currentItem = this.currentViewItem, max: number = this.maxRendered) {
        this.visibleSet = null;
        this.visibleSet = this.dir.getIntervalAround(currentItem, max, this.maxPrevItems);
        return this.visibleSet;
    }

    // Could also add in full directory load information here
    public imgLoaded(evt) {
        let img = evt.target;
        console.log("Img Loaded", img.naturalHeight, img.naturalWidth, img);
    }

    public imgClicked(evt) {
        this.clickedItem.emit({dir: this.dir, item: evt});
    }
}

