import {OnInit, Component, EventEmitter, Input, Output, HostListener} from '@angular/core';
import {MediaContainer} from './directory';

@Component({
    selector: 'contented-view',
    templateUrl: './contented_view.ng.html'
})
export class ContentedViewCmp {

    @Input() container: MediaContainer;

    @Input() forceWidth: number;
    @Input() forceHeight: number;

    public maxWidth: number;
    public maxHeight: number;

    // This calculation does _not_ work when using a dialog.  Fix?
    // Provide a custom width and height calculation option
    constructor() {
        this.calculateDimensions();
        console.log("Force width", this.forceWidth, this.forceHeight);
    }

    @HostListener('window:resize', ['$event'])
    public calculateDimensions() {
        // Probably should just set it via dom calculation of the actual parent
        // container?  Maybe?  but then I actually DO want scroll in some cases.
        if (this.forceWidth > 0) {
            let width = window.innerWidth; // document.body.clientWidth;
            this.maxWidth = width - 20 > 0 ? width - 20 : 640;
        } else {
            this.maxWidth = this.forceWidth;
        }

        if (this.forceHeight > 0) {
            // Probably need to do this off the current overall container
            let height = window.innerHeight; // document.body.clientHeight;
            this.maxHeight = height - 20 > 0 ? height - 20 : 480;
        } else {
            this.maxHeight = this.forceHeight;
        }
    }

}
