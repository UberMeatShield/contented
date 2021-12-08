import {OnInit, Component, EventEmitter, Input, Output, HostListener} from '@angular/core';
import {ContentedService} from './contented_service';
import {Container} from './container';
import {Media} from './media';
import {GlobalNavEvents} from './nav_events';

import * as _ from 'lodash';

@Component({
    selector: 'contented-nav',
    templateUrl: 'contented_nav.ng.html'
})
export class ContentedNavCmp implements OnInit {

    @Input() navEvts;

    constructor(public _contentedService: ContentedService) {

    }

    ngOnInit() {
        this.navEvts = this.navEvts || GlobalNavEvents.navEvts;
    }

    // On the document keypress events, listen for them (probably need to set them only to component somehow)
    @HostListener('document:keypress', ['$event'])
    public keyPress(evt: KeyboardEvent) {
        switch (evt.key) {
            case 'w':
                GlobalNavEvents.prevContainer();
                break;
            case 's':
                GlobalNavEvents.nextContainer();
                break;
            case 'a':
                GlobalNavEvents.prevMedia();
                break;
            case 'd':
                GlobalNavEvents.nextMedia();
                break;
            case 'e':
                GlobalNavEvents.viewFullScreen();
                break;
            case 'q':
                GlobalNavEvents.hideFullScreen();
                break;
            case 'f':
                GlobalNavEvents.loadMoreMedia();
                break;
            case 'x':
                GlobalNavEvents.saveMedia();
                break;
            default:
                break;
        }
    }
}

