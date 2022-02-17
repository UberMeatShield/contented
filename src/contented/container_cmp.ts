import {Subscription} from 'rxjs';
import {OnInit, OnDestroy, Component, EventEmitter, Input, Output, HostListener} from '@angular/core';
import {ContentedService} from './contented_service';

import {Container} from './container';
import {Media} from './media';
import {GlobalNavEvents, NavTypes} from './nav_events';
import * as _ from 'lodash';

@Component({
    selector: 'container-cmp',
    templateUrl: 'container.ng.html'
})
export class ContainerCmp implements OnInit, OnDestroy {

    @Input() container: Container;
    @Input() active: boolean = false;
    @Input() previewWidth: number;
    @Input() previewHeight: number;

    @Input() maxRendered: number = 8; // Default setting for how many should be visible at any given time
    @Input() maxPrevItems: number = 2; // When scrolling through a cnt, how many previous items should be visible

    @Output() clickedItem: EventEmitter<any> = new EventEmitter<any>();

    // @Output clickEvt: EventEmitter<any>;
    public visibleSet: Array<Media>; // The currently visible set of items from in the container
    public sub: Subscription;

    constructor(public _contentedService: ContentedService) {

    }

    public ngOnInit() {
        this.sub = GlobalNavEvents.navEvts.subscribe(evt => {
            if (this.active) {
                // console.log("Container Event found", this.container.name, evt);
                switch (evt.action) {
                    case NavTypes.NEXT_MEDIA:
                        console.log("Next in container");
                        this.nextMedia();
                        break;
                    case NavTypes.PREV_MEDIA:
                        console.log("Prev in container");
                        this.prevMedia();
                        break;
                    case NavTypes.SAVE_MEDIA:
                        console.log("Save the currently selected media");
                        this.saveMedia();
                        break;
                    default:
                        break;
                }
            }
        });
    }

    public ngOnDestroy() {
        this.sub.unsubscribe();
    }

    public saveMedia() {
        this._contentedService.download(this.container, this.container.rowIdx);
    }

    public nextMedia() {
        let mediaList = this.container.getContentList() || [];
        if (this.container.rowIdx < mediaList.length) {
            this.container.rowIdx++;
            if (this.container.rowIdx === mediaList.length) {
                GlobalNavEvents.nextContainer();
            } else {
                GlobalNavEvents.selectMedia(this.container.getCurrentMedia(), this.container);
            }
        }
    }

    public prevMedia() {
         if (this.container.rowIdx > 0) {
             this.container.rowIdx--;
             GlobalNavEvents.selectMedia(this.container.getCurrentMedia(), this.container);
         } else {
             GlobalNavEvents.prevContainer();
         }
    }

    public getVisibleSet(currentItem: Media = null, max: number = this.maxRendered) {
        let media: Media = currentItem || this.container.getCurrentMedia();
        this.visibleSet = this.container.getIntervalAround(media, max, this.maxPrevItems);
        return this.visibleSet;
    }

    // Could also add in full container load information here
    public imgLoaded(evt) {
        let img = evt.target;
        //console.log("Img Loaded", img.naturalHeight, img.naturalWidth, img);
    }

    public clickMedia(media: Media) {
        // Little strange on the selection
        this.container.rowIdx = _.findIndex(this.container.contents, {id: media.id});

        GlobalNavEvents.selectMedia(media, this.container);
        GlobalNavEvents.viewFullScreen(media);

        // Just here in case we want to override what happens on a click
        this.clickedItem.emit({cnt: this.container, media: media});
    }
}

