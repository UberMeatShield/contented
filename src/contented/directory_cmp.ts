import {OnInit, Component, EventEmitter, Input, Output, HostListener} from '@angular/core';
import {ContentedService} from './contented_service';

import {Directory} from './directory';

@Component({
    selector: 'directory-cmp',
    templateUrl: 'directory.ng.html'
})
export class DirectoryCmp implements OnInit {

    @Input() dir: Directory;
    @Input() previewWidth: number;
    @Input() previewHeight: number;
    @Input() currentViewItem: string;

    // @Output clickEvt: EventEmitter<any>;

    constructor() {

    }

    public ngOnInit() {
        console.log("Directory Component loading up");
    }

    public imgLoaded(evt) {
        let img = evt.target;
        console.log("Img Loaded", img.naturalHeight, img.naturalWidth, img);
    }
}

