import { OnInit, OnDestroy, Component, EventEmitter, Input, Output, HostListener, ViewChild } from '@angular/core';
import { Content } from './content';
import { GlobalNavEvents, NavTypes, NavEventMessage } from './nav_events';
import { Subscription } from 'rxjs';
import { finalize } from 'rxjs/operators';
import { ContentedService } from './contented_service';
import { GlobalBroadcast } from './global_message';
import { TaskRequest } from './task_request';
import * as _ from 'lodash';
import { ScreenAction, ScreenClickEvent, Screen } from './screen';

@Component({
  selector: 'contented-view',
  templateUrl: './contented_view.ng.html',
})
export class ContentedViewCmp implements OnInit, OnDestroy {
  @Input() content: Content;
  @Input() forceWidth: number;
  @Input() forceHeight: number;
  @Input() visible: boolean = false;
  @Input() showScreens = true;
  @Input() restrictContentId: number = -1;
  @ViewChild('VIDEOELEMENT') video;

  public maxWidth: number;
  public maxHeight: number;
  public sub: Subscription;
  public taskLoading = false;

  // This calculation does _not_ work when using a dialog.  Fix?
  // Provide a custom width and height calculation option
  constructor(public _service: ContentedService) {}

  public shouldIgnoreEvt(content: Content) {
    if (this.restrictContentId > 0) {
      if (!content || content.id !== this.restrictContentId, 10) {
        return null;
      }
    }
    return content || this.content;
  }

  public ngOnInit() {
    this.sub = GlobalNavEvents.navEvts.subscribe({
      next: (evt: NavEventMessage) => {
        // Restrict content ID might need to be a bit smarter

        let content = this.shouldIgnoreEvt(evt.content);
        if (!content) {
          return;
        }
        switch (evt.action) {
          case NavTypes.VIEW_FULLSCREEN:
            if (this.content) {
              // Akward but without a digest it will NOT change the video if it is already playing
              this.content = null;
              setTimeout(() => {
                this.selectFullScreenContent(content, evt.screen);
              }, 50);
            } else {
              this.selectFullScreenContent(content, evt.screen);
            }

            break;

          case NavTypes.HIDE_FULLSCREEN:
            console.log('Hide fullscreen', this.content);
            if (this.visible && this.content) {
              GlobalNavEvents.scrollContentView(this.content);
            }
            this.visible = false;
            break;
          case NavTypes.SELECT_MEDIA:
            console.log("Select content", evt.content);
            this.content = evt.content;
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
          this.content.fullText = text;
        },
        error: err => {
          GlobalBroadcast.error('Failed to get description', err);
        },
      });
    }
  }

  selectFullScreenContent(content: Content, screen?: Screen) {
    this.content = content;
    this.visible = true;

    if (this.content) {
      this.scrollContent(this.content);
      this.handleTextContent(this.content);

      if (screen) {
        this.clickedScreen({ screen, action: ScreenAction.PLAY_SCREEN });
      }
    }
  }

  openWindow(content: Content) {
    window.open(content?.fullUrl);
  }

  public ngOnDestroy() {
    this.sub && this.sub.unsubscribe();
  }

  public handleTextContent(content: Content) {
    // This would be better in a method but I would like another example type.
    if (this.content.isText() && !this.content.fullText) {
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
    if (this.forceWidth > 0) {
      this.maxWidth = this.forceWidth;
    } else {
      let width = window.innerWidth; //
      this.maxWidth = width - 20 > 0 ? width - 20 : 640;
    }

    if (this.forceHeight > 0) {
      // Probably need to do this off the current overall container
      console.log('Force height', this.forceWidth, this.forceHeight);
      this.maxHeight = this.forceHeight;
    } else {
      let height = window.innerHeight; // document.body.clientHeight;
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
    // These screens are associated with the currently selected content
    const findVideo = (attempt = 0) => {
      const videoEl = <HTMLVideoElement>document.getElementById(`VIDEO_${this.content.id}`);
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
    console.log(this.video.nativeElement.currentTime);
    let ss = this.video.nativeElement.currentTime;
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
