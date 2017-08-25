import {OnInit, Component, EventEmitter, Input, Output} from '@angular/core';

@Component({
    selector: 'contented-cmp',
    template: 'contented.ng.html'
})
export class ContentedCmp implements OnInit {

    constructor() {

    }

    public ngOnInit() {
        console.log("Contented comp is alive.");
    }
}
