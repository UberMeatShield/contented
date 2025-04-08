import { OnInit, Component, Input, HostListener, OnDestroy, ViewChild } from '@angular/core';
import { Content } from './content';
import { ContentedService } from './contented_service';

import { MatMenuTrigger } from '@angular/material/menu';
import { finalize } from 'rxjs/operators';
import { GlobalBroadcast } from './global_message';
import { GlobalNavEvents, NavEventMessage, NavTypes } from './nav_events';
import { Subscription } from 'rxjs';
import { Container, getFavorites } from './container';

import _ from 'lodash';
import { getWindowSizes } from './common';

@Component({
  selector: 'favorites-cmp',
  templateUrl: './favorites.ng.html',
})
export class FavoritesCmp implements OnInit, OnDestroy {
  @Input() container: Container | undefined;
  @Input() previewWidth: number = 480;
  @Input() previewHeight: number = 480;
  @Input() maxVisible: number = 16;
  @Input() visible: boolean = false;
  @Input() monitorFavorites: boolean = true;

  @ViewChild(MatMenuTrigger) contextMenu!: MatMenuTrigger;

  public contextMenuPosition = { x: '0px', y: '0px' };
  public sub: Subscription | undefined;
  public maxWidth: number | undefined;
  public maxHeight: number | undefined;
  public loading: boolean = false;
  public error: string | undefined;
  public active: boolean = false;

  constructor(public _service: ContentedService) {
    this.container = this.container || getFavorites();
  }

  onContextMenu(event: MouseEvent, content: Content) {
    event.preventDefault();
    if (!this.contextMenu) {
      return;
    }
    this.contextMenuPosition.x = event.clientX + 'px';
    this.contextMenuPosition.y = event.clientY + 'px';
    this.contextMenu.menuData = { content: content };
    this.contextMenu.menu?.focusFirstItem('mouse');
    this.contextMenu.openMenu();
  }

  public ngOnInit() {
    this.calculateDimensions();

    this.sub = GlobalNavEvents.navEvts.subscribe({
      next: (evt: NavEventMessage) => {
        // This container is not active but it should be monitoring favorites
        if (!evt.content) return;
        
        switch (evt.action) {
          case NavTypes.FAVORITE_MEDIA:
            this.handleFavorite(evt.content);
            break;
          case NavTypes.REMOVE_FAVORITE:
            this.removeFavorite(evt.content);
            break;
          case NavTypes.TOGGLE_DUPLICATE:
            this.handleToggleDuplicate(evt.content);
            break;
          case NavTypes.TOGGLE_FAVORITE_VISIBILITY:
            this.visible = !this.visible;
            if (this.container) {
              this.container.visible = this.visible;
            }
            break;
        }
      },
    });
  }

  public ngOnDestroy() {
    this.sub?.unsubscribe();
  }

  /**
   * This might be worth a full component with different behaviors (needs a user model to properly handle favorites)
   * @param content
   */
  public handleFavorite(content: Content) {
    if (!this.container) {
      return;
    }
    let idx = _.findIndex(this.container.contents, { id: content.id });
    if (idx >= 0) {
      _.remove(this.container.contents, { id: content.id });
    } else {
      this.container.addContents([content]);
      this.container.total = this.container.contents.length;
    }
  }

  public removeFavorite(content: Content) {
    if (!this.container) {
      return;
    }
    let idx = _.findIndex(this.container.contents, { id: content.id });
    if (idx >= 0) {
      _.remove(this.container.contents, { id: content.id });
    }
  }

  /**
   *
   * @param content
   */
  public toggleDuplicate(content: Content) {
    GlobalNavEvents.toggleDuplicate(content);
  }

  public handleToggleDuplicate(content: Content) {
    content.duplicate = !content.duplicate;
    this._service.saveContent(content).subscribe({
      next: (updated: Content) => {
        content.duplicate = updated.duplicate;
      },
      error: err => {
        GlobalBroadcast.error(err);
      },
    });
  }

  public clickContent(content: Content) {
    GlobalNavEvents.viewFullScreen(content);
  }

  // TODO: Being called abusively in the constructor rather than on page resize events
  @HostListener('window:resize', ['$event'])
  public calculateDimensions() {
    // This should be based on the container not the window
    // but unfortunately we call it before it is in the dom and visible
    // so there is a load operation order issue to solve.  Maybe afterViewInit would work?
    const {width, height} = getWindowSizes();

    // 120 is right if the top nav is hidden, could calculate that it is out of view for the height of things
    this.previewWidth = width / this.maxVisible - 12;
    this.previewHeight = height / 6;
  }
}
