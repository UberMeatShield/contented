import { Component, OnInit, OnDestroy, Input, HostListener, ViewChild } from '@angular/core';
import { Container } from './container';
import { Content } from './content';
import { MatMenuTrigger } from '@angular/material/menu';
import { ActivatedRoute, Router } from '@angular/router';
import { ContentedService } from './contented_service';
import { GlobalNavEvents, NavTypes, NavEventMessage } from './nav_events';
import { Subscription } from 'rxjs';
import { finalize } from 'rxjs/operators';
import { GlobalBroadcast } from './global_message';
import * as _ from 'lodash';
import { initializeDefaults } from './utils';
import { getWindowSize } from './common';

@Component({
    selector: 'favorites-cmp',
    templateUrl: './favorites.ng.html',
    standalone: false
})
export class FavoritesCmp implements OnInit, OnDestroy {
  @Input() container!: Container;
  @Input() previewWidth!: number;
  @Input() previewHeight!: number;
  @Input() searchCount = 4;
  @Input() showToggle = true;
  @Input() maxItems: number = 16; // For calculating display dimensions

  @ViewChild(MatMenuTrigger) contextMenu!: MatMenuTrigger;
  public contextMenuPosition = { x: '0px', y: '0px' };
  public contextMenuContent: Content | undefined;
  public sub: Subscription | undefined;
  public maxWidth: number | undefined;
  public maxHeight: number | undefined;
  public containerVisible = true;

  constructor(
    public _contentedService: ContentedService,
    public route: ActivatedRoute,
    public router: Router
  ) {
    // Initialize all properties
    initializeDefaults(this, {
      sub: new Subscription(),
      maxWidth: 0,
      maxHeight: 0,
      previewWidth: 200,
      previewHeight: 200
    });
  }

  public ngOnInit() {
    this.calculateDimensions();
    this.sub = GlobalNavEvents.navEvts.subscribe({
      next: (evt: NavEventMessage) => {
        if (evt.action === NavTypes.TOGGLE_FAVORITE_VISIBILITY) {
          this.toggleVisible();
        }
        if (evt.action === NavTypes.FAVORITE_MEDIA) {
          this.handleFavorite(evt.content!);
        }
        if (evt.action === NavTypes.REMOVE_FAVORITE) {
          this.removeFavorite(evt.content!);
        }
        if (evt.action === NavTypes.TOGGLE_DUPLICATE) {
          this.handleToggleDuplicate(evt.content!);
        }
      },
    });
  }

  onContextMenu(event: MouseEvent, content: Content) {
    event.preventDefault();
    this.contextMenuPosition.x = event.clientX + 'px';
    this.contextMenuPosition.y = event.clientY + 'px';
    this.contextMenuContent = content;
    if (this.contextMenu && this.contextMenu.menu) {
      this.contextMenu.menu.focusFirstItem('mouse');
      this.contextMenu.openMenu();
    }
  }

  public ngOnDestroy() {
    this.sub?.unsubscribe();
  }

  /**
   * This might be worth a full component with different behaviors (needs a user model to properly handle favorites)
   * @param content
   */
  public handleFavorite(content: Content) {
    let idx = _.findIndex(this.container.contents, { id: content.id });
    if (idx >= 0) {
      _.remove(this.container.contents, { id: content.id });
    } else {
      this.container.addContents([content]);
      this.container.total = this.container.contents.length;
    }
  }

  public removeFavorite(content: Content) {
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
    this._contentedService.saveContent(content).subscribe({
      next: (updated: any) => {
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
    const {width, height} = getWindowSize();

    // 120 is right if the top nav is hidden, could calculate that it is out of view for the height of things
    this.previewWidth = width / this.maxItems - 12;
    this.previewHeight = height / 6;
  }

  public toggleVisible() {
    this.containerVisible = !this.containerVisible;
    this.container.visible = this.containerVisible;
  }
}
