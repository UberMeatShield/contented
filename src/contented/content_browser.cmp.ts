import { Subscription } from 'rxjs';
import { OnInit, OnDestroy, Component, Input, HostListener } from '@angular/core';
import { ContentedService } from './contented_service';
import { Container, LoadStates } from './container';
import { Content } from './content';
import { finalize } from 'rxjs/operators';

import { ActivatedRoute, Router, ParamMap } from '@angular/router';
import { GlobalNavEvents, NavTypes, NavEventMessage } from './nav_events';

import { GlobalBroadcast } from './global_message';
import _ from 'lodash';

@Component({
  selector: 'content-browser',
  templateUrl: 'content_browser.ng.html',
  standalone: false,
})
export class ContentBrowserCmp implements OnInit, OnDestroy {
  @Input() maxVisible: number = 2; // How many of the loaded containers should we be viewing
  @Input() perRow: number = 6; // How many items per row should we be viewing
  @Input() rowIdx: number = 0; // Which row (content item) are we on
  @Input() idx: number = 0; // Which item within the container are we viewing

  public loading: boolean = false;
  public emptyMessage: string | undefined;

  // TODO: Remove this listener
  public fullScreen: boolean = false; // Should we view fullscreen the current item
  public containers: Array<Container> = []; // Current set of visible containers
  public allCnts: Array<Container> = []; // All the containers we have loaded
  public sub: Subscription | undefined;

  constructor(
    public _contentedService: ContentedService,
    public route: ActivatedRoute,
    public router: Router
  ) {}

  public ngOnInit() {
    // Need to load content if the idx is greater than content loaded (n times potentially)
    this.route.paramMap.pipe().subscribe({
      next: (res: ParamMap) => {
        this.setPosition(
          res.get('idx') ? parseInt(res.get('idx') || '0', 10) : this.idx,
          res.get('rowIdx') ? parseInt(res.get('rowIdx') || '0', 10) : this.rowIdx
        );
      },
      error: err => {
        GlobalBroadcast.error('Failed to set position using route', err);
      },
    });
    this.setupEvtListener();

    if (_.isEmpty(this.allCnts)) {
      this.loadContainers(); // Do this after the param map load potentially
    }
  }

  ngOnDestroy() {
    if (this.sub) {
      this.sub.unsubscribe();
    }
  }

  public setupEvtListener() {
    this.sub = GlobalNavEvents.navEvts.subscribe({
      next: (evt: NavEventMessage) => {
        switch (evt.action) {
          case NavTypes.NEXT_CONTAINER:
            this.next();
            break;
          case NavTypes.PREV_CONTAINER:
            this.prev();
            break;
          case NavTypes.LOAD_MORE:
            this.loadMore();
            break;
          case NavTypes.SELECT_MEDIA:
            if (evt.content) {
              this.selectedContent(evt.content, evt.cnt);
            }
            break;
          case NavTypes.SELECT_CONTAINER:
            this.selectContainer(evt.cnt);
            break;
          default:
            break;
        }
      },
    });
  }

  public loadMore() {
    let visible = this.getVisibleContainers();
    this.loadMoreInDir(visible[0]);
  }

  // Mostly for tests since testing full routing params is a god damn pain.
  public setPosition(idx: number, rowIdx: number) {
    this.idx = idx;
    this.rowIdx = rowIdx;
  }

  public loadContainers() {
    this.loading = true;
    this._contentedService
      .getContainers()
      .pipe(
        finalize(() => {
          this.loading = false;
        })
      )
      .subscribe({
        next: res => {
          this.previewResults(res.results);
        },
        error: err => {
          console.error(err);
        },
      });
  }

  public loadMoreInDir(cnt: Container) {
    if (cnt.count < cnt.total && cnt.loadState === LoadStates.Partial) {
      this.loading = true;
      this._contentedService
        .loadMoreInDir(cnt)
        .pipe(
          finalize(() => {
            this.loading = false;
          })
        )
        .subscribe({
          next: res => {
            this.cntResults(cnt, res);
          },
          error: err => {
            console.error(err);
          },
        });
    }
  }

  public cntResults(cnt: Container, response: { total: number; results: Array<Content> }) {
    // console.log("Results loading, what is in the results?", response);
    cnt.addContents(response?.results);
  }

  public reset() {
    this.idx = 0;
    this.allCnts = [];
    this.emptyMessage = undefined;
  }

  public getVisibleContainers() {
    if (this.allCnts) {
      let start = this.idx < this.allCnts.length ? this.idx : this.allCnts.length - 1;
      let end = start + this.maxVisible <= this.allCnts.length ? start + this.maxVisible : this.allCnts.length;

      // Only loads if cnt.loadState = LoadStates.NotLoaded
      let currCnt = this.getCurrentContainer();
      let cnts = this.allCnts.slice(start, end);

      _.each(cnts, (cnt, _idx) => {
        let obs = this._contentedService.initialLoad(cnt);
        if (obs) {
          obs.subscribe({
            next: content => {
              if (cnt == currCnt) {
                const currentContent = cnt.getContent();
                if (currentContent) {
                  GlobalNavEvents.selectContent(currentContent, cnt);
                }
              }
            },
            error: err => console.error,
          });
        }
      });
      return cnts;
    }
    return [];
  }

  public selectContainer(cnt: Container | undefined) {
    if (!cnt) {
      return;
    }
    let idx = _.findIndex(this.allCnts, { id: cnt.id });
    if (idx >= 0) {
      this.idx = idx;
      console.log('This idx', this.idx);
      this.selectionEvt();
    }
  }

  // Ensure the route is set and if we moved containers it should show
  // what has been selected.
  public selectionEvt() {
    let cnt = this.getCurrentContainer();
    if (cnt) {
      console.log('Selected container ID', cnt.id);
      const currentContent = cnt.getContent();
      if (currentContent) {
        GlobalNavEvents.selectContent(currentContent, cnt);
      }
      this.updateRoute();
    }
  }

  public getCurrentContainer(): Container | undefined {
    if (this.idx < this.allCnts.length && this.idx >= 0) {
      return this.allCnts[this.idx];
    }
    return undefined;
  }

  public updateRoute() {
    let cnt = this.allCnts[this.idx];
    this.router.navigate([`/ui/browse/${this.idx}/${cnt.rowIdx}`]);
  }

  public next(selectFirst: boolean = true) {
    if (this.allCnts && this.idx + 1 < this.allCnts.length) {
      this.idx++;
      this.selectionEvt();
    }
  }

  public prev(selectLast: boolean = false) {
    if (this.idx > 0) {
      this.idx--;
      this.selectionEvt();
    }
  }

  public previewResults(containers: Array<Container>) {
    // console.log("Results returned from the preview results.", containers);
    this.allCnts = containers || [];
    if (_.isEmpty(containers)) {
      this.emptyMessage = 'No Directories found, did you load the DB?';
    } else {
      // Maybe just read the current param from the route
      this.loadView(this.idx, this.rowIdx, true);
    }
    return this.allCnts;
  }

  public fullLoadDir(cnt: Container) {
    this._contentedService.fullLoadDir(cnt).subscribe({
      next: (loadedCnt: Container) => {
        console.log('Loading all content in directory', loadedCnt.id);
      },
      error: err => {
        console.error('Failed to load', err);
      },
    });
  }

  public loadView(idx: number, rowIdx: number, triggerSelect: boolean = false) {
    let currDir = this.getCurrentContainer();
    if (currDir && rowIdx >= currDir.total) {
      rowIdx = 0;
    }
    this.idx = idx;
    if (currDir) {
      currDir.rowIdx = rowIdx;
    }

    // This handles the case where we need to fully load a container to reach the row
    if (currDir && rowIdx >= currDir.count) {
      this.fullLoadDir(currDir);
    } else if (triggerSelect) {
      let cnt = this.getCurrentContainer();
      if (cnt) {
        GlobalNavEvents.selectContent(cnt.getContent(), cnt);
      }
    }
  }

  // Could probably move this into a saner location
  public selectedContent(content: Content | undefined, cnt: Container | undefined) {
    if (!cnt) {
      return;
    }
    //console.log("Click event, change currently selected indexes, container etc", content, cnt);
    let idx = _.findIndex(this.allCnts, { id: cnt ? cnt.id : -1 });
    if (idx >= 0) {
      this.idx = idx;
      this.rowIdx = cnt.rowIdx;
      this.updateRoute();
    } else {
      console.error('Should not be able to click an item we cannot find.', cnt, content);
    }
  }
}
