import { Subscription } from 'rxjs';
import { OnInit, Component, EventEmitter, Input, Output, HostListener } from '@angular/core';
import { ContentedService } from './contented_service';
import { finalize } from 'rxjs/operators';

import { Screen, ScreenAction, ScreenClickEvent } from './screen';
import { GlobalBroadcast } from './global_message';
import * as _ from 'lodash';

@Component({
  selector: 'screens-cmp',
  templateUrl: 'screens.ng.html',
})
export class ScreensCmp implements OnInit {
  @Input() contentId: number;
  @Input() screens: Array<Screen>;
  @Input() previewWidth: number = 480;
  @Input() previewHeight: number = 480;

  // Allow something to force specify the values
  @Input() containerWidth: number = null;
  @Input() containerHeight: number = null;

  @Output() screensLoaded: EventEmitter<Array<Screen>> = new EventEmitter<Array<Screen>>();
  @Output() screenClick: EventEmitter<ScreenClickEvent> = new EventEmitter<ScreenClickEvent>();

  // TODO: These are not used
  /*
    @Input() maxRendered: number = 8; // Default setting for how many should be visible at any given time
    @Input() maxPrevItems: number = 2; // When scrolling through a cnt, how many previous items should be visible
    */
  public loading: boolean = false;

  // @Output clickEvt: EventEmitter<any>;
  public sub: Subscription;

  constructor(public _contentedService: ContentedService) {}

  public ngOnInit() {
    this.calculateDimensions();

    if (this.contentId) {
      this.loading = true;
      this._contentedService
        .getScreens(this.contentId)
        .pipe(
          finalize(() => {
            this.loading = false;
          })
        )
        .subscribe({
          next: (res: { total: number; results: Screen[] }) => {
            // Could emit an event for the screens loading and listen so it updates the content
            this.screens = res.results;
            this.calculateDimensions();
            this.screensLoaded.emit(this.screens);
          },
          error: err => {
            console.error(err);
          },
        });
    }
  }

  public clickContent(screen: Screen) {
    // Potentially this should just submit the global nav event
    this.screenClick.emit({
      screen: screen,
      screens: this.screens,
      action: ScreenAction.VIEW,
    });
  }

  public clickTime(screen: Screen, evt: Event) {
    evt.preventDefault();
    evt.stopPropagation();
    console.log('Screen time Click information exists on the screen?', screen);

    this.screenClick.emit({
      screen: screen,
      screens: this.screens,
      action: ScreenAction.PLAY_SCREEN,
    });
  }

  // Should grab the content dimensions
  @HostListener('window:resize', ['$event'])
  public calculateDimensions() {
    // TODO: Should this base the screen sizing on dom container vs the overall window?
    if (!this.screens) {
      return;
    }

    let perRow = Math.ceil((this.screens?.length || 0) / 2);
    let width = !window['jasmine'] ? window.innerWidth : 800;
    if (this.containerWidth) {
      width = this.containerWidth;
    }
    let height = !window['jasmine'] ? window.innerHeight : 800;
    if (this.containerHeight) {
      height = this.containerHeight;
    }

    // This should be based on the total number of screens?
    this.previewWidth = Math.ceil(width / perRow);
    this.previewHeight = Math.floor(height / 4); // We want 2 rows
  }
}
