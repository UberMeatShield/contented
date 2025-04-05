import { Subscription } from 'rxjs';
import { OnInit, OnDestroy, Component, EventEmitter, Input, Output, HostListener, ViewChild } from '@angular/core';
import { MatMenuTrigger } from '@angular/material/menu';
import { ContentedService } from './contented_service';

import { Container } from './container';
import { Content } from './content';
import { GlobalNavEvents, NavTypes, NavEventMessage } from './nav_events';
import * as _ from 'lodash';

@Component({
    selector: 'container-cmp',
    templateUrl: 'container.ng.html',
    standalone: false
})
export class ContainerCmp implements OnInit, OnDestroy {
  @Input() container: Container;
  @Input() active: boolean = false;
  @Input() maxWidth: number = 0;
  @Input() maxHeight: number = 0;
  @ViewChild(MatMenuTrigger) contextMenu: MatMenuTrigger;

  @Input() maxRendered: number = 8; // Default setting for how many should be visible at any given time
  @Input() maxPrevItems: number = 2; // When scrolling through a cnt, how many previous items should be visible
  @Input() monitorFavorites: boolean = false;

  @Output() clickedItem: EventEmitter<any> = new EventEmitter<any>();

  // @Output clickEvt: EventEmitter<any>;
  public previewWidth: number = 0;
  public previewHeight: number = 0;
  public visibleSet: Array<Content>; // The currently visible set of items from in the container
  public sub: Subscription;
  public contextMenuPosition = { x: '0px', y: '0px' };

  constructor(public _contentedService: ContentedService) {}

  public ngOnInit() {
    this.sub = GlobalNavEvents.navEvts.subscribe({
      next: (evt: NavEventMessage) => {
        if (this.active) {
          // console.log("Container Event found", this.container.name, evt);
          switch (evt.action) {
            case NavTypes.NEXT_MEDIA:
              console.log('Next in container');
              this.nextContent();
              break;
            case NavTypes.PREV_MEDIA:
              console.log('Prev in container');
              this.prevContent();
              break;
            case NavTypes.SAVE_MEDIA:
              console.log('Save the currently selected content');
              this.saveContent();
              break;
            case NavTypes.TOGGLE_FAVORITE:
              this.toggleFavorite();
              break;
            case NavTypes.SCROLL_MEDIA_INTO_VIEW:
              this.scrollContent(evt.content);
              break;
            default:
              break;
          }
        }
      },
    });

    this.calculateDimensions();
  }

  onContextMenu(event: MouseEvent, content: Content) {
    event.preventDefault();
    this.contextMenuPosition.x = event.clientX + 'px';
    this.contextMenuPosition.y = event.clientY + 'px';
    this.contextMenu.menuData = { content: content };
    this.contextMenu.menu.focusFirstItem('mouse');
    this.contextMenu.openMenu();
  }

  public scrollContent(content: Content) {
    _.delay(() => {
      let id = `preview_${content.id}`;
      let el = document.getElementById(id);

      if (el) {
        el.scrollIntoView(true);
        window.scrollBy(0, -30);
      }
    }, 20);
  }

  addFavorite(content: Content) {
    GlobalNavEvents.favoriteContent(content);
  }

  public toggleDuplicate(content: Content) {
    GlobalNavEvents.toggleDuplicate(content);
  }

  /**
   * (keypress 't') If there is a current media element selected then we should toggle it.
   */
  public toggleFavorite() {
    if (this.container.rowIdx >= 0 && this.container.rowIdx < this.container?.contents?.length) {
      const content = this.container.contents[this.container.rowIdx];
      GlobalNavEvents.favoriteContent(content);
    }
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
    this.container.rowIdx = _.findIndex(this.container.contents, {
      id: content.id,
    });

    GlobalNavEvents.selectContent(content, this.container);
    GlobalNavEvents.viewFullScreen(content);

    // Just here in case we want to override what happens on a click
    this.clickedItem.emit({ cnt: this.container, content: content });
  }

  // TODO: Being called abusively in the constructor rather than on page resize events
  @HostListener('window:resize', ['$event'])
  public calculateDimensions() {
    // This should be based on the container not the window
    // but unfortunately we call it before it is in the dom and visible
    // so there is a load operation order issue to solve.  Maybe afterViewInit would work?
    let width = this.maxWidth || !window['jasmine'] ? window.innerWidth : 800;
    let height = this.maxHeight || !window['jasmine'] ? window.innerHeight : 800;

    // 120 is right if the top nav is hidden, could calculate that it is out of view for the height of things
    // when doing navigation. Potentially the sizing could be done in the container and the max with provided
    // to the container (maxWidth, maxHeight)
    const halfVisible = Math.ceil(this.maxRendered / 2);
    this.previewWidth = Math.ceil(width / halfVisible - 12);
    this.previewHeight = Math.ceil((height - 160) / 2);
  }
}
