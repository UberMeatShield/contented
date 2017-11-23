import {OnInit, Component, EventEmitter, Input, Output, HostListener} from '@angular/core';

@Component({
    selector: 'contented-view',
    templateUrl: './contented_view.ng.html'
})
export class ContentedViewCmp {
    @Input() viewUrl: string = '';

    constructor() {

    }

    public maxWidth: number;
    public maxHeight: number;
    public calculateDimensions() {

        // Probably need to do this off the current overall container
        let width = document.body.clientWidth;
        let height = document.body.clientHeight;

        this.maxWidth = width - 20;
        this.maxHeight = height - 20;
    }

}
