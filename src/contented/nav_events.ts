import { EventEmitter } from '@angular/core';
import { Content } from './content';
import { Container } from './container';
import { Screen } from './screen';


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

export interface NavEventMessage {
  action: NavTypes;
  content: Content | undefined;
  cnt: Container | undefined;
  screen?: Screen | undefined;
}

export class NavEvents {
  // Subscribe to the navEvts in order to act on comand in the app
  public navEvts: EventEmitter<NavEventMessage>;

  constructor() {
    this.navEvts = new EventEmitter<NavEventMessage>();
  }

  prevContainer() {
    this.navEvts.emit({
      action: NavTypes.PREV_CONTAINER,
      content: undefined,
      cnt: undefined,
    });
  }

  nextContainer() {
    this.navEvts.emit({
      action: NavTypes.NEXT_CONTAINER,
      content: undefined,
      cnt: undefined,
    });
  }

  // TODO: Need a select container event I guess
  selectContainer(container: Container) {
    this.navEvts.emit({
      action: NavTypes.SELECT_CONTAINER,
      content: undefined,
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
      content: undefined,
    });
  }

  prevContent(container: Container = null) {
    this.navEvts.emit({
      action: NavTypes.PREV_MEDIA,
      cnt: container,
      content: undefined,
    });
  }

  viewFullScreen(content: Content = null, screen?: Screen, container?: Container) {
    this.navEvts.emit({
      action: NavTypes.VIEW_FULLSCREEN,
      content: content,
      cnt: container,
      screen: screen,
    });
  }

  hideFullScreen() {
    // Require no content
    this.navEvts.emit({
      action: NavTypes.HIDE_FULLSCREEN,
      cnt: undefined,
      content: undefined,
    });
  }

  loadMoreContent(container: Container = null) {
    this.navEvts.emit({
      action: NavTypes.LOAD_MORE,
      cnt: container,
      content: undefined,
    });
  }

  // Determine if this should require a content element
  saveContent(content: Content = null) {
    this.navEvts.emit({
      action: NavTypes.SAVE_MEDIA,
      content: content,
      cnt: undefined,
    });
  }

  scrollContentView(content: Content = null) {
    this.navEvts.emit({
      action: NavTypes.SCROLL_MEDIA_INTO_VIEW,
      content: content,
      cnt: undefined,
    });
  }
}

export const GlobalNavEvents = new NavEvents();
