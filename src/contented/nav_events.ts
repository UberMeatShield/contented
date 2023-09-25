import {EventEmitter} from '@angular/core';
import {Content} from './content';
import {Container} from './container';

export enum NavTypes {
    NEXT_CONTAINER,
    PREV_CONTAINER,
    SELECT_MEDIA,
    SELECT_CONTAINER,
    NEXT_MEDIA,
    PREV_MEDIA,
    VIEW_FULLSCREEN,
    HIDE_FULLSCREEN,
    LOAD_MORE,
    SAVE_MEDIA,
    SCROLL_MEDIA_INTO_VIEW,
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

    // TODO: Need a select container event I guess
    selectContainer(container: Container) {
        this.navEvts.emit({
            action: NavTypes.SELECT_CONTAINER,
            cnt: container,
        });
    }
    

    selectContent(content: Content, container: Container) {
        this.navEvts.emit({
            action: NavTypes.SELECT_MEDIA,
            content: content,
            cnt: container,
        });
    }

    nextContent(container: Container = null) {
        this.navEvts.emit({
            action: NavTypes.NEXT_MEDIA,
            cnt: container,
        });
    }

    prevContent(container: Container = null) {
        this.navEvts.emit({
            action: NavTypes.PREV_MEDIA ,
            cnt: container,
        });
    }

    viewFullScreen(content: Content = null) {
        this.navEvts.emit({
            action: NavTypes.VIEW_FULLSCREEN,
            content: content,
        });
    }

    hideFullScreen() {
        // Require no content
        this.navEvts.emit({
            action: NavTypes.HIDE_FULLSCREEN,
        });
    }

    loadMoreContent(container: Container = null) {
        this.navEvts.emit({
            action: NavTypes.LOAD_MORE,
            cnt: container,
        });
    }

    // Determine if this should require a content element
    saveContent(content: Content = null) {
        this.navEvts.emit({
            action: NavTypes.SAVE_MEDIA,
            content: content,
        });
    }

    scrollContentView(content: Content = null) {
        this.navEvts.emit({
            action: NavTypes.SCROLL_MEDIA_INTO_VIEW,
            content: content,
        });
    }
}

export const GlobalNavEvents = new NavEvents();
