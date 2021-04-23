import {OnInit, Component, EventEmitter, Input, Output, HostListener} from '@angular/core';
import {ImgContainer} from './directory';

@Component({
    selector: 'contented-view',
    templateUrl: './contented_view.ng.html'
})
export class ContentedViewCmp {

    @Input() container: ImgContainer;
    public maxWidth: number;
    public maxHeight: number;

    constructor() {
        this.calculateDimensions();
    }

    public calculateDimensions() {
        // Probably need to do this off the current overall container
        let width = document.body.clientWidth;
        let height = document.body.clientHeight;

        // Just adds more min rendering size
        this.maxWidth = width - 20 > 0 ? width - 20 : 640;
        this.maxHeight = height - 20 > 0 ? height - 20 : 480;
    }

}
