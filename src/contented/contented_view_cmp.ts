import {OnInit, OnDestroy, Component, EventEmitter, Input, Output, HostListener} from '@angular/core';
import {Media} from './media';
import {GlobalNavEvents, NavTypes} from './nav_events';
import {Subscription} from 'rxjs';

@Component({
    selector: 'contented-view',
    templateUrl: './contented_view.ng.html'
})
export class ContentedViewCmp implements OnInit, OnDestroy {

    @Input() media: Media;
    @Input() forceWidth: number;
    @Input() forceHeight: number;
    @Input() visible: boolean = false;

    public maxWidth: number;
    public maxHeight: number;
    public sub: Subscription;

    // This calculation does _not_ work when using a dialog.  Fix?
    // Provide a custom width and height calculation option
    constructor() {
    }

    public ngOnInit() {
        this.calculateDimensions();
        this.sub = GlobalNavEvents.navEvts.subscribe(evt => {
            switch(evt.action) {
                case NavTypes.VIEW_FULLSCREEN:
                    console.log("Viewscreen show media");
                    this.visible = true;
                    this.media = evt.media || this.media;
                    break;
                case NavTypes.HIDE_FULLSCREEN:
                    console.log("Viewscreen hide media");
                    this.visible = false;
                    break;
                case NavTypes.SELECT_MEDIA:
                    this.media = evt.media;
                    break;
            }
            console.log("Listen for the fullscreen");
        }) ;
    }

    public ngOnDestroy() {
        this.sub.unsubscribe();
    }

    @HostListener('window:resize', ['$event'])
    public calculateDimensions() {
        // Probably should just set it via dom calculation of the actual parent
        // container?  Maybe?  but then I actually DO want scroll in some cases.
        if (this.forceWidth > 0) {
            this.maxWidth = this.forceWidth;
        } else {
            let width = window.innerWidth; // document.body.clientWidth;
            this.maxWidth = width - 20 > 0 ? width - 20 : 640;
        }

        if (this.forceHeight > 0) {
            // Probably need to do this off the current overall container
            console.log("Force height", this.forceWidth, this.forceHeight);
            this.maxHeight = this.forceHeight;
        } else {
            let height = window.innerHeight; // document.body.clientHeight;
            this.maxHeight = height - 20 > 0 ? height - 20 : 480;
        }
    }

}
