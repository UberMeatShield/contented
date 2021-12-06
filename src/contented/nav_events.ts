import {EventEmitter} from '@angular/core';
import {Media} from './media';
import {Container} from './container';

export enum NavTypes {
    NEXT_CONTAINER,
    PREV_CONTAINER,
    SELECT_MEDIA,
    NEXT_MEDIA,
    PREV_MEDIA,
    VIEW_FULLSCREEN,
    HIDE_FULLSCREEN,
    LOAD_MORE,
    SAVE_MEDIA,
}

export class NavEvents {

    // Subscribe to the navEvts in order to act on comand in the app
    public navEvts: EventEmitter<any>;

    constructor() {
        this.navEvts = new EventEmitter<any>();
    }

    prevContainer() {
        this.navEvts.emit({
            action: NavTypes.PREV_CONTAINER
        });
    }

    nextContainer() {
        this.navEvts.emit({
            action: NavTypes.NEXT_CONTAINER
        });
    }

    selectMedia(media: Media, container: Container) {
        this.navEvts.emit({
            action: NavTypes.SELECT_MEDIA,
            media: media,
            cnt: container,
        });
    }

    nextMedia(container: Container = null) {
        this.navEvts.emit({
            action: NavTypes.NEXT_MEDIA,
            cnt: container,
        });
    }

    prevMedia(container: Container = null) {
        this.navEvts.emit({
            action: NavTypes.PREV_MEDIA ,
            cnt: container,
        });
    }

    viewFullScreen(media: Media = null) {
        this.navEvts.emit({
            action: NavTypes.VIEW_FULLSCREEN,
            cnt: media,
        });
    }

    hideFullScreen() {
        // Require no media
        this.navEvts.emit({
            action: NavTypes.HIDE_FULLSCREEN,
        });
    }

    loadMoreMedia(container: Container = null) {
        this.navEvts.emit({
            action: NavTypes.LOAD_MORE,
            cnt: container,
        });
    }

    // Determine if this should require a media element
    saveMedia(media: Media = null) {
        this.navEvts.emit({
            action: NavTypes.SAVE_MEDIA,
            media: media,
        });
    }
}

export const GlobalNavEvents = new NavEvents;
