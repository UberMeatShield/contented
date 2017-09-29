import {OnInit, Component, EventEmitter, Input, Output, HostListener} from '@angular/core';

@Component({
    selector: 'contented-view',
    template: require('./contented_view.ng.html')
})
export class ContentedViewCmp {
    @Input() viewUrl: string = '';

    constructor() {

    }
}
