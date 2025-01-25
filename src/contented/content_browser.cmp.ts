import { Subscription } from 'rxjs';
import { OnInit, OnDestroy, Component, Input, HostListener } from '@angular/core';
import { ContentedService } from './contented_service';
import { Container } from './container';
import { Content } from './content';
import { finalize } from 'rxjs/operators';

import { ActivatedRoute, Router, ParamMap } from '@angular/router';
import { GlobalNavEvents, NavTypes, NavEventMessage } from './nav_events';

import { GlobalBroadcast } from './global_message';
import * as _ from 'lodash';

@Component({
  selector: 'content-browser',
  templateUrl: 'content_browser.ng.html',
})
export class ContentBrowserCmp implements OnInit, OnDestroy {
  @Input() maxVisible: number = 2; // How many of the loaded containers should we be viewing
  @Input() rowIdx: number = 0; // Which row (content item) are we on
  @Input() idx: number = 0; // Which item within the container are we viewing

  public loading: boolean = false;
  public emptyMessage = null;
  public previewWidth: number = 200; // Based on current client page sizes, scale the preview images natually
  public previewHeight: number = 200; // height for the previews ^

  // TODO: Remove this listener
  public fullScreen: boolean = false; // Should we view fullscreen the current item
  public containers: Array<Container>; // Current set of visible containers
  public allCnts: Array<Container>; // All the containers we have loaded
  public sub: Subscription;
  public favoritesContainer: Container;

  constructor(
    public _contentedService: ContentedService,
    public route: ActivatedRoute,
    public router: Router
  ) {
    this.favoritesContainer = this.getFavoritesContainer();
  }

  public ngOnInit() {
    // Need to load content if the idx is greater than content loaded (n times potentially)
    this.route.paramMap.pipe().subscribe({
      next: (res: ParamMap) => {
        this.setPosition(
          res.get('idx') ? parseInt(res.get('idx'), 10) : this.idx,
          res.get('rowIdx') ? parseInt(res.get('rowIdx'), 10) : 0
        );
      },
      error: err => {
        GlobalBroadcast.error('Failed to set position using route', err);
      },
    });
    this.setupEvtListener();
    this.calculateDimensions();
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
            this.selectedContent(evt.content, evt.cnt);
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

  public getFavoritesContainer() {
    if (!this.favoritesContainer) {
      this.favoritesContainer = new Container({
        id: 'favorites',
        name: 'Favorites',
        previewUrl: 'https://placehold.co/200x200',
        contents: [],
        total: 0,
        count: 0,
        rowIdx: 0,
      });
    }
    return this.favoritesContainer;
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
    // This is being changed to just load more content up
    if (cnt.count < cnt.total && !this.loading) {
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

  public cntResults(cnt: Container, response) {
    // console.log("Results loading, what is in the results?", response);
    cnt.addContents(cnt.buildImgs(response));
  }

  public reset() {
    this.idx = 0;
    this.allCnts = [];
    this.emptyMessage = null;
  }

  public getVisibleContainers() {
    if (this.allCnts) {
      let start = this.idx < this.allCnts.length ? this.idx : this.allCnts.length - 1;
      let end = start + this.maxVisible <= this.allCnts.length ? start + this.maxVisible : this.allCnts.length;
      // Only loads if cnt.loadState = LoadStates.NotLoaded
      let currCnt = this.getCurrentContainer();
      let cnts = this.allCnts.slice(start, end);
      _.each(cnts, (cnt, idx) => {
        let obs = this._contentedService.initialLoad(cnt);
        if (obs) {
          obs.subscribe({
            next: content => {
              if (cnt == currCnt) {
                GlobalNavEvents.selectContent(cnt.getContent(), cnt);
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

  public selectContainer(cnt: Container) {
    let idx = _.findIndex(this.allCnts, { id: cnt.id });
    console.log('Selected container', cnt.id, idx);
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
    console.log('Selected container ID', cnt.id);
    GlobalNavEvents.selectContent(cnt.getContent(), cnt);
    this.updateRoute();
  }

  public getCurrentContainer() {
    if (this.idx < this.allCnts.length && this.idx >= 0) {
      return this.allCnts[this.idx];
    }
    return null;
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

  // TODO: Being called abusively in the constructor rather than on page resize events
  @HostListener('window:resize', ['$event'])
  public calculateDimensions() {
    // This should be based on the container not the window
    // but unfortunately we call it before it is in the dom and visible
    // so there is a load operation order issue to solve.  Maybe afterViewInit would work?
    let width = !window['jasmine'] ? window.innerWidth : 800;
    let height = !window['jasmine'] ? window.innerHeight : 800;

    // 120 is right if the top nav is hidden, could calculate that it is out of view for the height of things
    // when doing navigation.
    this.previewWidth = (width / 4) - 12;
    this.previewHeight = (height - 160) / this.maxVisible;
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
        console.log('Fully loaded up the container', loadedCnt.id);
        GlobalNavEvents.selectContent(loadedCnt.getContent(), loadedCnt);
      },
      error: err => {
        console.error('Failed to load', err);
      },
    });
  }

  public loadView(idx: number, rowIdx: number, triggerSelect: boolean = false) {
    let currDir = this.getCurrentContainer();
    if (rowIdx >= currDir.total) {
      rowIdx = 0;
    }
    this.idx = idx;
    currDir.rowIdx = rowIdx;

    console.log('LoadView', currDir, rowIdx, triggerSelect);
    // This handles the case where we need to fully load a container to reach the row
    if (rowIdx >= currDir.count) {
      this.fullLoadDir(currDir);
    } else if (triggerSelect) {
      let cnt = this.getCurrentContainer();
      GlobalNavEvents.selectContent(cnt.getContent(), cnt);
    }
  }

  // Could probably move this into a saner location
  public selectedContent(content: Content, cnt: Container) {
    //console.log("Click event, change currently selected indexes, container etc", content, cnt);
    let idx = _.findIndex(this.allCnts, { id: cnt ? cnt.id : '-1' });
    if (idx >= 0) {
      this.idx = idx;
      this.rowIdx = cnt.rowIdx;
      this.updateRoute();
    } else {
      console.error('Should not be able to click an item we cannot find.', cnt, content);
    }
  }
}
