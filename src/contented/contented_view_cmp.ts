import {OnInit, Component, EventEmitter, Input, Output, HostListener} from '@angular/core';
import {Media} from './media';

@Component({
    selector: 'contented-view',
    templateUrl: './contented_view.ng.html'
})
export class ContentedViewCmp {

    @Input() container: Media;

    @Input() forceWidth: number;
    @Input() forceHeight: number;

    public maxWidth: number;
    public maxHeight: number;

    // This calculation does _not_ work when using a dialog.  Fix?
    // Provide a custom width and height calculation option
    constructor() {
        this.calculateDimensions();
    }

    @HostListener('window:resize', ['$event'])
    public calculateDimensions() {
        // Probably should just set it via dom calculation of the actual parent
        // container?  Maybe?  but then I actually DO want scroll in some cases.
        if (this.forceWidth > 0) {
            console.log("Force width", this.forceWidth, this.forceHeight);
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
