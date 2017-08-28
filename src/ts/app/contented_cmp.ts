import {OnInit, Component, EventEmitter, Input, Output} from '@angular/core';
import {ContentedService} from './contented_service';

@Component({
    selector: 'contented-main',
    template: require('./contented.ng.html')
})
export class ContentedCmp implements OnInit {

    constructor(public _contentedService: ContentedService) {

    }

    public ngOnInit() {
        console.log("Contented comp is alive.");
        this._contentedService.getPreview().subscribe(
            console.log,
            console.error
        );
    }
}
