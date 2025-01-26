import { OnInit, Component, Input, HostListener, OnDestroy } from '@angular/core';
import { Content } from './content';
import { ContentedService } from './contented_service';
import { ActivatedRoute, Router, ParamMap } from '@angular/router';

import { finalize } from 'rxjs/operators';
import { GlobalBroadcast } from './global_message';
import { GlobalNavEvents, NavEventMessage, NavTypes } from './nav_events';
import { Subscription } from 'rxjs';
import { Container } from './container';

import _ from 'lodash';

@Component({
  selector: 'favorites-cmp',
  templateUrl: './favorites.ng.html',
})
export class FavoritesCmp implements OnInit, OnDestroy {
  @Input() container: Container;
  @Input() previewWidth: number;
  @Input() previewHeight: number;
  @Input() maxVisible: number = 16;
  @Input() visible: boolean = false;
  @Input() monitorFavorites: boolean = true;

  public sub: Subscription;
  public maxWidth: number;
  public maxHeight: number;
  public loading: boolean = false;
  public error = null;
  public active: boolean = false;

  constructor(
    public _service: ContentedService,
    public route: ActivatedRoute,
    public router: Router
  ) {}

  public ngOnInit() {
    this.container =
      this.container ||
      new Container({
        id: 'favorites',
        name: 'Favorites',
        previewUrl: 'https://placehold.co/200x200',
        contents: [],
        total: 0,
        count: 0,
        rowIdx: 0,
      });
    this.calculateDimensions();

    this.sub = GlobalNavEvents.navEvts.subscribe({
      next: (evt: NavEventMessage) => {
        // This container is not active but it should be monitoring favorites
        if (this.monitorFavorites && evt.action === NavTypes.FAVORITE_MEDIA) {
          this.handleFavorite(evt.content);
        }
      },
    });
  }

  public ngOnDestroy() {
    this.sub?.unsubscribe();
  }

  /**
   * This might be worth a full component with different behaviors
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

  public clickContent(content: Content) {
    console.log('Clicked content', content);
  }

  // TODO: Being called abusively in the constructor rather than on page resize events
  @HostListener('window:resize', ['$event'])
  public calculateDimensions() {
    // This should be based on the container not the window
    // but unfortunately we call it before it is in the dom and visible
    // so there is a load operation order issue to solve.  Maybe afterViewInit would work?
    let width = !window['jasmine'] ? window.innerWidth : 800;
    let height = !window['jasmine'] ? window.innerHeight : 800;

    // 120 is right if the top nav is hidden, could calculate that it is out of view for the height of things
    this.previewWidth = width / this.maxVisible - 12;
  }
}
