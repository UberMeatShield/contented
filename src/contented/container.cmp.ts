import {Subscription} from 'rxjs';
import {OnInit, OnDestroy, Component, EventEmitter, Input, Output} from '@angular/core';
import {ContentedService} from './contented_service';

import {Container} from './container';
import {Content} from './content';
import {GlobalNavEvents, NavTypes, NavEventMessage} from './nav_events';
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
    public visibleSet: Array<Content>; // The currently visible set of items from in the container
    public sub: Subscription;

    constructor(public _contentedService: ContentedService) {

    }

    public ngOnInit() {
        this.sub = GlobalNavEvents.navEvts.subscribe({
            next: (evt: NavEventMessage) => {
                if (this.active) {
                    // console.log("Container Event found", this.container.name, evt);
                    switch (evt.action) {
                        case NavTypes.NEXT_MEDIA:
                            console.log("Next in container");
                            this.nextContent();
                            break;
                        case NavTypes.PREV_MEDIA:
                            console.log("Prev in container");
                            this.prevContent();
                            break;
                        case NavTypes.SAVE_MEDIA:
                            console.log("Save the currently selected content");
                            this.saveContent();
                            break;
                        case NavTypes.SCROLL_MEDIA_INTO_VIEW:
                            this.scrollContent(evt.content);
                            break;
                        default:
                            break;
                    }
                }
            }
        });
    }

    public scrollContent(content: Content) {
        _.delay(() => {
            let id = `preview_${content.id}`;
            let el = document.getElementById(id)

            if (el) {
                el.scrollIntoView(true);
                window.scrollBy(0, -30);
            }
        }, 20);
    }

    public ngOnDestroy() {
        if (this.sub) {
            this.sub.unsubscribe();
        }
    }

    public saveContent() {
        this._contentedService.download(this.container, this.container.rowIdx);
    }

    public nextContent() {
        let contentList = this.container.getContentList() || [];
        if (this.container.rowIdx < contentList.length) {
            this.container.rowIdx++;
            if (this.container.rowIdx === contentList.length) {
                GlobalNavEvents.nextContainer();
            } else {
                GlobalNavEvents.selectContent(this.container.getCurrentContent(), this.container);
            }
        }
    }

    public prevContent() {
         if (this.container.rowIdx > 0) {
             this.container.rowIdx--;
             GlobalNavEvents.selectContent(this.container.getCurrentContent(), this.container);
         } else {
             GlobalNavEvents.prevContainer();
         }
    }

    public getVisibleSet(currentItem: Content = null, max: number = this.maxRendered) {
        let content: Content = currentItem || this.container.getCurrentContent();
        this.visibleSet = this.container.getIntervalAround(content, max, this.maxPrevItems);
        return this.visibleSet;
    }

    // Could also add in full container load information here
    public imgLoaded(evt) {
        let img = evt.target;
        //console.log("Img Loaded", img.naturalHeight, img.naturalWidth, img);
    }

    public clickContent(content: Content) {
        // Little strange on the selection
        this.container.rowIdx = _.findIndex(this.container.contents, {id: content.id});

        GlobalNavEvents.selectContent(content, this.container);
        GlobalNavEvents.viewFullScreen(content);

        // Just here in case we want to override what happens on a click
        this.clickedItem.emit({cnt: this.container, content: content});
    }
}

