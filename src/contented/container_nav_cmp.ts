import {Subscription} from 'rxjs';
import {OnInit, OnDestroy, Component, EventEmitter, Input, Output, HostListener} from '@angular/core';
import {ContentedService} from './contented_service';
import {Container} from './container';
import {Media} from './media';
import {finalize, switchMap} from 'rxjs/operators';

import {ActivatedRoute, Router, ParamMap} from '@angular/router';
import {GlobalNavEvents, NavTypes} from './nav_events';

import * as _ from 'lodash';

@Component({
    selector: 'container-nav',
    templateUrl: 'container_nav.ng.html'
})
export class ContainerNavCmp implements OnInit, OnDestroy {

    // This is actually required
    @Input() cnt: Container;

    // Do we actually care?
    @Input() totalContainers: number = 0;

    // current view Item should be something you trigger per directory (move view ?)
    public currentMedia: Media;

    // idx and current view item might be better as a top level nav / hover should be allowed?
    @Input() active: boolean = false; // Is our container active

    // rowIdx should be independently controlled for each directory
    @Output() navEvt: EventEmitter<any> = new EventEmitter<any>();
    @Input() rowIdx: number = 0; // Which media item is selected
    @Input() idx: number = 0; // What is our index compared to other containers

    private sub: Subscription;

    constructor() {

    }

    public ngOnInit() {
        this.sub = GlobalNavEvents.navEvts.subscribe(evt => {
            console.log("Did we get this select event", evt);
            if (evt.action == NavTypes.SELECT_MEDIA) {
                this.currentMedia = evt.media;
            }
        });
    }

    public ngOnDestroy() {
        this.sub.unsubscribe();
    }

    fullLoadDir(cnt: Container) {
        console.log('This button should work in the nav');
    }
}
