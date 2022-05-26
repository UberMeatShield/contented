import {OnInit, OnDestroy, Component, EventEmitter, Input, Output, HostListener} from '@angular/core';
import {Media} from './media';
import {ContentedService} from './contented_service';
import {ActivatedRoute, Router, ParamMap} from '@angular/router';

import {finalize} from 'rxjs/operators';

@Component({
    selector: 'media-view',
    templateUrl: './media_view.ng.html'
})
export class MediaViewCmp implements OnInit {

    @Input() media: Media;
    @Input() forceWidth: number;
    @Input() forceHeight: number;
    @Input() visible: boolean = false;

    public maxWidth: number;
    public maxHeight: number;
    public loading: boolean = false;
    public error = null;

    constructor(public _service: ContentedService,  public route: ActivatedRoute, public router: Router) {

    }

    public ngOnInit() {
        this.route.paramMap.pipe().subscribe(
            (res: ParamMap) => {
                let mediaID = res.get("id")
                if (mediaID) {
                    this.loadMedia(mediaID);
                }
            }, err => { console.error(err); }
        )
    }

    public loadMedia(mediaID: string) {
        this.loading = true;
        this._service.getMedia(mediaID).pipe(
            finalize(() => { this.loading = false; })
        ).subscribe(
            (m: Media) => {
                this.media = m;
            }, err => {
                console.error("Failed to load media", err);
                this.error = "Failed to find mediaID" + err;
            }
        )
    }
}
