import { forkJoin, Subscription } from 'rxjs';
import {
  OnInit,
  OnDestroy,
  AfterViewInit,
  Component,
  EventEmitter,
  Input,
  Output,
  HostListener,
  ViewChild,
  Inject,
} from '@angular/core';
import { ContentedService } from './contented_service';
import { Content } from './content';
import { Container } from './container';
import { Screen, ScreenAction } from './screen';
import { GlobalNavEvents, NavTypes } from './nav_events';
import { ActivatedRoute, Router, ParamMap } from '@angular/router';
import { FormBuilder, NgForm, FormControl, FormGroup } from '@angular/forms';

import { PageEvent as PageEvent } from '@angular/material/paginator';
import { MatDialog as MatDialog, MatDialogConfig as MatDialogConfig, MAT_DIALOG_DATA } from '@angular/material/dialog';
import * as _ from 'lodash';

@Component({
  selector: 'video-preview-cmp',
  templateUrl: './video_preview.ng.html',
})
export class VideoPreviewCmp implements OnInit {
  // Is this preview the selected one
  @Input() selected = false;
  @Input() content?: Content;

  // Used to show that only a certain number are visible on the page at a given time
  // which is used to sort of set a height on the components.
  @Input() maxVisible = 2;
  @Input() inlineView = false;

  // TODO: Make this a saner calculation
  public previewWidth = 480;
  public previewHeight = 480;
  public screenWidth = 960;

  constructor(public dialog: MatDialog) {}

  public screensLoaded(screens: Array<Screen>) {
    //console.log("Screens loaded", screens);
    if (this.content) {
      this.content.screens = this.content.screens || [];
      _.each(screens, screen => {
        if (this.content && this.content.id == screen.content_id) {
          this.content.screens.push(screen);
        }
      });
    }
  }

  public ngOnInit() {
    this.calculateDimensions();
  }

  // A little awkward and needs to be fixed (attempt to do a lookup)
  public fullView(mc: Content, screen?: Screen) {
    // This needs to be fixed to not scroll up
    GlobalNavEvents.selectContent(mc, null);

    // Just makes sure the selection event doesn't race condition the scroll
    // into view event.  So the click triggers, scrolls and then we scroll to
    // the fullscreen element.

    _.delay(() => {
      GlobalNavEvents.viewFullScreen(mc, screen);
    }, 50);
  }

  // Rather than window I should probably make it the containing dom element?
  @HostListener('window:resize', ['$event'])
  public calculateDimensions() {
    let width = !window['jasmine'] ? window.innerWidth : 800;
    let height = !window['jasmine'] ? window.innerHeight : 800;

    this.previewWidth = width / 5;
    this.previewHeight = height / this.maxVisible - 41;

    // screenHeight is just calculated on the component previewHeight * 2
    this.screenWidth = width - this.previewWidth - 200; // Fudge factor
  }

  public screenEvt(evt) {
    console.log('Screen Evt', evt);
    if (evt.action === ScreenAction.PLAY_SCREEN) {
      return this.fullView(this.content, evt.screen);
    }

    if (evt.action === ScreenAction.VIEW) {
      const dialogRef = this.dialog.open(ScreenDialog, {
        data: { screen: evt.screen, screens: evt.screens },
        width: '90%',
        height: '100%',
        maxWidth: '100vw',
        maxHeight: '100vh',
      });
      dialogRef.afterClosed().subscribe({
        next: result => {
          console.log('Closing the Dialog on VideoPreview', result);
        },
      });
    }
  }

  imgClicked(mc: Content) {
    this.fullView(mc);
  }
}

// This just doesn't seem like a great approach :(
@Component({
  selector: 'screen-dialog',
  templateUrl: 'screen_dialog.ng.html',
})
export class ScreenDialog implements AfterViewInit {
  public screen: Screen;
  public screens: Array<Screen>;

  public forceHeight: number;
  public forceWidth: number;
  public sizeCalculated: boolean = false;
  @ViewChild('ScreensContent', { static: true }) screenContent;

  constructor(
    @Inject(MAT_DIALOG_DATA) public data,
    public _service: ContentedService
  ) {
    this.screen = data.screen;
    this.screens = data.screens;
  }

  ngAfterViewInit() {
    console.log('Search content is:', this.screenContent);
    setTimeout(() => {
      let el = this.screenContent.nativeElement;
      if (el) {
        console.log('Element', el, el.offsetWidth, el.offsetHeight);
        this.forceHeight = el.offsetHeight - 100;
        this.forceWidth = el.offsetWidth - 100;
      }
      this.sizeCalculated = true;
    }, 100);
  }

  idx() {
    if (this.screens && this.screen) {
      return _.findIndex(this.screens, { id: this.screen.id });
    }
    return -1;
  }

  next() {
    let i = this.idx();
    if (i < this.screens.length - 1) {
      this.screen = this.screens[i + 1];
    }
  }

  prev() {
    let i = this.idx();
    if (i - 1 >= 0) {
      this.screen = this.screens[i - 1];
    }
  }
}
