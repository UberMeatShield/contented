import {Subscription} from 'rxjs';
import {OnInit, OnDestroy, Component, EventEmitter, Input, Output, HostListener} from '@angular/core';
import {ContentedService} from './contented_service';
import {finalize, switchMap} from 'rxjs/operators';

import {Screen} from './screen';
import {GlobalNavEvents, NavTypes} from './nav_events';
import * as _ from 'lodash';

@Component({
    selector: 'screens-cmp',
    templateUrl: 'screens.ng.html'
})
export class ScreensCmp implements OnInit {

    @Input() mediaId: string;
    @Input() screens: Array<Screen>;
    @Input() previewWidth: number = 480;
    @Input() previewHeight: number = 480;

    @Input() maxRendered: number = 8; // Default setting for how many should be visible at any given time
    @Input() maxPrevItems: number = 2; // When scrolling through a cnt, how many previous items should be visible

    @Output() clickedItem: EventEmitter<any> = new EventEmitter<any>();
    public loading: boolean = false;

    // @Output clickEvt: EventEmitter<any>;
    public sub: Subscription;

    constructor(public _contentedService: ContentedService) {

    }

    public ngOnInit() {
        if (this.mediaId) {
            this.loading = true;
            this._contentedService.getScreens(this.mediaId).pipe(
                finalize(() => { this.loading = false; })
            ).subscribe(
                (screens: Array<Screen>) => {
                    console.log("No screens", screens);
                    this.screens = screens;
                    this.calculateDimensions();
                }, err => {
                    console.error(err);
                }
            );
        }
    }

    public clickMedia(screen: Screen) {
        // Just here in case we want to override what happens on a click
        this.clickedItem.emit({screen: screen});
    }

    @HostListener('window:resize', ['$event'])
    public calculateDimensions() {

        // TODO: Should this base the screen sizing on dom container vs the overall window?
        let perRow = this.screens ? (this.screens.length / 2) : 6;
        let width = !window['jasmine'] ? window.innerWidth : 800;
        let height = !window['jasmine'] ? window.innerHeight : 800;

        // This should be based on the total number of screens?
        this.previewWidth = ((width - 41) / perRow);
        this.previewHeight = (((height / 2) - 41)/ 2);
    }
}

