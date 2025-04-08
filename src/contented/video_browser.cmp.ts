import { Subscription } from 'rxjs';
import { finalize, debounceTime, distinctUntilChanged } from 'rxjs/operators';

import { OnInit, OnDestroy, Component, ViewChild, Input } from '@angular/core';
import { ContentedService, ContentSearchSchema, PageResponse } from './contented_service';
import { Content, Tag, VSCodeChange } from './content';
import { Container } from './container';
import { GlobalNavEvents, NavTypes, NavEventMessage } from './nav_events';
import { ActivatedRoute, Router, ParamMap, Params } from '@angular/router';
import { FormBuilder, FormControl, FormGroup } from '@angular/forms';
import { GlobalBroadcast } from './global_message';

import { PageEvent as PageEvent } from '@angular/material/paginator';
import * as _ from 'lodash';
import { MatMenuTrigger } from '@angular/material/menu';

@Component({
  selector: 'video-browser-cmp',
  templateUrl: './video_browser.ng.html',
})
export class VideoBrowserCmp implements OnInit, OnDestroy {
  // Route needs to exist
  // Take in the search text route param
  // Debounce the search
  @ViewChild('searchForm', { static: true }) searchControl;
  @ViewChild(MatMenuTrigger) contextMenu: MatMenuTrigger;
  @Input() tags: Array<Tag>;
  throttleSearch: Subscription;

  searchText: string; // Initial value
  searchType = new FormControl('text');
  currentTextChange: VSCodeChange = { value: '', tags: [] };
  changedSearch: (evt: VSCodeChange) => void;

  options: FormGroup;
  fb: FormBuilder;

  public selectedContent: Content | undefined; // For keeping track of where we are in the page
  public selectedContainer: Container; // For filtering
  public content: Array<Content>;
  public containers: Array<Container>;

  // TODO: Make this a saner calculation
  public maxVisible = 3; // How many results show vertically
  public total = 0;
  public offset = 0; // Tracking where we are in the position
  public pageSize = 50;
  public loading: boolean = false;
  public sub: Subscription; // Listening for GlobalNavEvents

  constructor(
    public _contentedService: ContentedService,
    public route: ActivatedRoute,
    public router: Router,
    fb: FormBuilder
  ) {
    this.fb = fb;
  }

  public ngOnInit() {
    this.changedSearch = _.debounce((evt: VSCodeChange) => {
      // Do not change this.searchText it will re-assign the VS-Code editor in a
      // bad way and muck with the cursor.
      this.search(evt.value, 0, 50, this.getCntId(), evt.tags);
      this.currentTextChange = evt;
    }, 250);

    // This should also preserve the current page we have selected and restore it.
    this.resetForm();
    this.setupEvtListener();
    this.route.queryParams.pipe().subscribe({
      next: (res: Params) => {
        this.searchText = res['searchText'] || '';

        // Add in a param for container_id ?
        // this.search(this.searchText, this.offset, this.pageSize, this.getCntId());
        this.loadContainers();
      },
    });
  }

  public contextMenuPosition = { x: '0px', y: '0px' };
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

  addFavorite(content: Content) {
    GlobalNavEvents.favoriteContent(content);
  }

  public toggleDuplicate(content: Content) {
    GlobalNavEvents.toggleDuplicate(content);
  }

  ngOnDestroy() {
    if (this.sub) {
      this.sub.unsubscribe();
    }
  }

  public loadContainers() {
    this._contentedService.getContainers().subscribe({
      next: (cnts: PageResponse<Container>) => {
        this.containers = cnts.results || [];
      },
      error: err => {
        GlobalBroadcast.error('Containers could not load', err);
      },
    });
  }

  // This will listen to nav events.
  public setupEvtListener() {
    this.sub = GlobalNavEvents.navEvts.subscribe({
      next: (evt: NavEventMessage) => {
        switch (evt.action) {
          case NavTypes.NEXT_MEDIA:
            this.next();
            break;
          case NavTypes.PREV_MEDIA:
            this.prev();
            break;
          case NavTypes.HIDE_FULLSCREEN:
            // Scroll back into view
            if (this.selectedContent) {
              this.selectContent(this.selectedContent, this.selectedContainer);
            }
            break;
          case NavTypes.LOAD_MORE:
            // this.loadMore();
            // It might not be TOO abusive to override this and make it page next?
            break;
          case NavTypes.SELECT_MEDIA:
              this.selectContent(evt.content, evt.cnt || undefined);
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

  public selectContainer(cnt: Container | undefined) {
    if (!cnt) {
      return;
    }
    let offset = this.offset;
    if (cnt?.id !== this.selectedContainer?.id) {
      this.offset = 0;
    }
    this.selectedContainer = cnt;
    this.search(this.currentTextChange.value, this.offset, this.pageSize, this.getCntId());
  }

  public next() {
    // It should have a jump to scroll location for the currently selected item
    if (this.selectedContent && this.content) {
      let idx = _.findIndex(this.content, { id: this.selectedContent.id });
      if (idx + 1 < this.content.length) {
        let m = this.content[idx + 1];
        if (m.id != this.selectedContent.id) {
          GlobalNavEvents.selectContent(m, new Container({ id: m.container_id }));
        }
      } else if (this.offset + this.pageSize < this.total) {
        this.search(this.currentTextChange.value, this.offset + this.pageSize, this.pageSize, this.getCntId());
      }
    }
  }

  public prev() {
    if (this.selectedContent && this.content) {
      let idx = _.findIndex(this.content, { id: this.selectedContent.id });
      if (idx - 1 >= 0) {
        let m = this.content[idx - 1];
        if (m.id != this.selectedContent.id) {
          GlobalNavEvents.selectContent(m, new Container({ id: m.container_id }));
        }
      } else if (this.offset - this.pageSize >= 0) {
        this.search(this.currentTextChange.value, this.offset - this.pageSize, this.pageSize, this.getCntId());
      }
    }
  }

  public selectContent(content: Content | undefined, container: Container | undefined) {
    this.selectedContent = content;
    if (!content) {
      return;
    }
    let id = `view_content_${content?.id}`;
    let el = document.getElementById(id);

    // Might want to debounce this as well
    if (el) {
      el.scrollIntoView(true);
      window.scrollBy(0, -60);
    }
  }

  public resetForm(setupFilterEvents: boolean = false) {
    this.options = this.fb.group({
      searchType: this.searchType,
    });
    if (setupFilterEvents) {
      this.setupFilterEvts();
    }
  }

  public setupFilterEvts() {
    // Kicks off a search
    if (this.throttleSearch) {
      this.throttleSearch.unsubscribe();
    }
    this.throttleSearch = this.options.valueChanges
      .pipe(
        debounceTime(250),
        distinctUntilChanged()
        // Prevent bubble on keypress
      )
      .subscribe({
        next: formData => {
          console.log('Form data changing');
          // If the text changes do we reset the search offset etc.
          this.search(this.currentTextChange.value, 0, this.pageSize, this.getCntId(), this.currentTextChange.tags);
        },
        error: error => {
          GlobalBroadcast.error('Failed to search', error);
        },
      });
  }

  getValues() {
    return this.options.value;
  }

  pageEvt(evt: PageEvent) {
    console.log('Event', evt, this.currentTextChange.value);
    let offset = evt.pageIndex * evt.pageSize;
    let limit = evt.pageSize;
    this.search(this.currentTextChange.value, offset, limit, this.getCntId());
  }

  public getCntId(): string {
    return !!this.selectedContainer ? this.selectedContainer.id.toString() : "";
  }

  // TODO: Add in optional filter params like the container (filter by container in search?)
  public search(
    text: string = '',
    offset: number = 0,
    limit: number = 50,
    cntId: string = "",
    tags: Array<string> = []
  ) {
    console.log('Get the information from the input and search on it', text, offset, limit, cntId);

    this.selectedContent = undefined;
    this.content = [];
    this.loading = true;

    const cs = ContentSearchSchema.parse({
      text,
      cId: cntId,
      offset,
      limit,
      contentType: 'video',
      tags,
    });

    this._contentedService
      .searchContent(cs)
      .pipe(finalize(() => (this.loading = false)))
      .subscribe({
        next: res => {
          let content = res.results;
          let total = res.total || 0;

          this.offset = offset;
          this.content = content;
          this.total = total;

          if (content && content.length > 0) {
            let mc = content[0];
            GlobalNavEvents.selectContent(mc, new Container({ id: mc.container_id }));
          }
        },
        error: err => {
          GlobalBroadcast.error('Failed to search for video content.', err);
        },
      });
  }

  // This will have to be updated to actually work
  public getVisibleSet() {
    return this.content;
  }

  public getContainer(cId: string) {
    return _.find(this.containers, { id: cId });
  }
}
