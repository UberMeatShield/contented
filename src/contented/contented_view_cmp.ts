import {
  OnInit,
  OnDestroy,
  Component,
  EventEmitter,
  Input,
  Output,
  HostListener,
  ViewChild,
  ElementRef,
} from '@angular/core';
import { Content } from './content';
import { GlobalNavEvents, NavTypes, NavEventMessage } from './nav_events';
import { Subscription } from 'rxjs';
import { finalize } from 'rxjs/operators';
import { ContentedService } from './contented_service';
import { GlobalBroadcast } from './global_message';
import { TaskRequest } from './task_request';
import * as _ from 'lodash';
import { ScreenAction, ScreenClickEvent, Screen } from './screen';
import { getWindowSizes } from './common';

@Component({
  selector: 'contented-view',
  templateUrl: './contented_view.ng.html',
  standalone: false,
})
export class ContentedViewCmp implements OnInit, OnDestroy {
  @Input() content?: Content;
  @Input() forceWidth: number = 0;
  @Input() forceHeight: number = 0;
  @Input() visible: boolean = false;
  @Input() showScreens = true;
  @Input() restrictContentId: number = -1;
  @ViewChild('VIDEOELEMENT') video: ElementRef<HTMLVideoElement> | undefined;

  public maxWidth: number = 800;
  public maxHeight: number = 600;
  public sub: Subscription | undefined;
  public taskLoading = false;

  // This calculation does _not_ work when using a dialog.  Fix?
  // Provide a custom width and height calculation option
  constructor(public _service: ContentedService) {}

  public shouldIgnoreEvt(content: Content | undefined): boolean {
    if (this.restrictContentId > 0) {
      if (!content || content.id !== this.restrictContentId) {
        return true;
      }
    }
    return false;
  }

  public ngOnInit() {
    this.sub = GlobalNavEvents.navEvts.subscribe({
      next: (evt: NavEventMessage) => {
        // Restrict content ID might need to be a bit smarter
        if (this.shouldIgnoreEvt(evt.content)) {
          return;
        }

        const content = evt.content;
        switch (evt.action) {
          case NavTypes.VIEW_FULLSCREEN:
            if (content) {
              // Akward but without a digest it will NOT change the video if it is already playing
              this.content = content;
              setTimeout(() => {
                this.selectFullScreenContent(content, evt.screen);
              }, 50);
            }
            break;

          case NavTypes.HIDE_FULLSCREEN:
            if (this.visible && content) {
              GlobalNavEvents.scrollContentView(evt.content);
            }
            this.visible = false;
            break;
          case NavTypes.SELECT_MEDIA:
            this.content = content;
            break;
        }
      },
      // console.log("Listen for the fullscreen");
    });
    this.calculateDimensions();

    // I don't like this hack but from a visibility standpoint it kinda works.
    if (this.content && this.content.isText() && this.visible) {
      this._service.getTextContent(this.content).subscribe({
        next: (text: string) => {
          if (this.content) {
            this.content.fullText = text;
          }
        },
        error: err => {
          GlobalBroadcast.error('Failed to get description', err);
        },
      });
    }
  }

  public selectFullScreenContent(content: Content, screen?: Screen) {
    if (!content) return;

    this.content = content;
    this.visible = true;

    this.scrollContent(content);
    this.handleTextContent(content);

    if (screen) {
      this.clickedScreen({ screen, action: ScreenAction.PLAY_SCREEN });
    }
  }

  openWindow(content: Content) {
    if (!content) return;
    window.open(content.fullUrl);
  }

  public ngOnDestroy() {
    this.sub && this.sub.unsubscribe();
  }

  public handleTextContent(content: Content) {
    if (!content) return;

    // This would be better in a method but I would like another example type.
    if (content.isText() && !content.fullText) {
      this._service.getTextContent(content).subscribe({
        next: (text: string) => {
          content.fullText = text;
        },
        error: err => {
          GlobalBroadcast.error('Failed to handle text updates', err);
        },
      });
    }
  }

  @HostListener('window:resize', ['$event'])
  public calculateDimensions() {
    // Probably should just set it via dom calculation of the actual parent
    // container?  Maybe?  but then I actually DO want scroll in some cases.
    const { width, height } = getWindowSizes();
    if (this.forceWidth > 0) {
      this.maxWidth = this.forceWidth;
    } else {
      this.maxWidth = width - 20 > 0 ? width - 20 : 640;
    }

    if (this.forceHeight > 0) {
      // Probably need to do this off the current overall container
      this.maxHeight = this.forceHeight;
    } else {
      this.maxHeight = height - 20 > 0 ? height - 64 : 480;
    }
  }

  public scrollContent(content: Content) {
    _.delay(() => {
      let id = `MEDIA_VIEW_${this.restrictContentId}`;
      let el = document.getElementById(id);
      if (el) {
        el.scrollIntoView(true);
        window.scrollBy(0, -30);
      }
    }, 50);
  }

  public clickedScreen(evt: ScreenClickEvent, count: number = 0) {
    if (!this.content) return;

    // These screens are associated with the currently selected content
    const findVideo = (attempt = 0) => {
      const videoEl = <HTMLVideoElement>document.getElementById(`VIDEO_${this.content?.id}`);
      if (videoEl) {
        videoEl.currentTime = evt.screen?.parseSecondsFromScreen() || 0;
        videoEl.play();
        return;
      }

      if (attempt < 50) {
        // 50 attempts * 100ms = 5 seconds
        setTimeout(() => findVideo(attempt + 1), 100);
      }
    };
    findVideo();
  }

  // Kinda just need the ability to get the task info from the server
  screenshot(content: Content) {
    // Determine how to get the current video index, if not defined then just use the default
    console.log(this.video?.nativeElement?.currentTime);

    let ss = this.video?.nativeElement?.currentTime || 0;
    this.taskLoading = true;
    this._service
      .requestScreens(content, 1, ss)
      .pipe(finalize(() => (this.taskLoading = false)))
      .subscribe({
        next: (task: TaskRequest) => {
          console.log('Queued a new task for getting a screen', task);
        },
        error: err => {
          GlobalBroadcast.error(`Could not get a screen at this time ${ss} and ${content.id}`, err);
        },
      });
  }
}
