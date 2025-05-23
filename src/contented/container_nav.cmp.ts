import { Subscription } from 'rxjs';
import { OnInit, OnDestroy, Component, EventEmitter, Input, Output } from '@angular/core';
import { ContentedService } from './contented_service';
import { Container, LoadStates } from './container';
import { Content } from './content';

import { GlobalNavEvents, NavTypes, NavEventMessage } from './nav_events';
import { FormGroup, FormBuilder, FormControl, Validators } from '@angular/forms';
import { GlobalBroadcast } from './global_message';

import * as _ from 'lodash';

@Component({
  selector: 'container-nav',
  templateUrl: 'container_nav.ng.html',
  standalone: false,
})
export class ContainerNavCmp implements OnInit, OnDestroy {
  // This is actually required
  @Input() cnt!: Container;

  // Do we actually care?
  @Input() totalContainers: number = 0;

  // current view Item should be something you trigger per directory (move view ?)
  public currentContent: Content | undefined;
  public ContainerLoadStates = LoadStates;

  // idx and current view item might be better as a top level nav / hover should be allowed?
  @Input() active: boolean = false; // Is our container active

  // rowIdx should be independently controlled for each directory
  @Output() navEvt: EventEmitter<any> = new EventEmitter<any>();
  @Input() rowIdx: number = 0; // Which content item is selected
  @Input() idx: number = 0; // What is our index compared to other containers

  private sub: Subscription | undefined;

  public navForm: FormGroup | undefined;
  public idxControl: FormControl<number | null>;

  constructor(
    public fb: FormBuilder,
    public _contentedService: ContentedService
  ) {
    this.idxControl = new FormControl(this.idx || this.cnt?.rowIdx || 0, Validators.required);
  }

  public ngOnInit() {
    this.navForm = this.fb.group({
      idxControl: this.idxControl,
    });

    this.sub = GlobalNavEvents.navEvts.subscribe({
      next: (evt: NavEventMessage) => {
        // console.log("Select Media", evt, evt.cnt == this.cnt, evt.action, evt.content);
        if (evt.action == NavTypes.SELECT_MEDIA && evt.cnt == this.cnt && evt.content) {
          //console.log("Container Nav found select content", evt, evt.cnt.name);
          this.currentContent = evt.content;
          this.idxControl?.setValue(this.cnt.rowIdx);
        }
      },
    });
    // The select event can trigger BEFORE a render loop so on a new render
    // ensure we at least get our current content (should be correct given the rowIdx)
    if (this.cnt) {
      this.currentContent = this.cnt.getContent();
    }

    this.navForm.get('idxControl')?.valueChanges.subscribe({
      next: idx => {
        if (idx != this.cnt.rowIdx) {
          let content = this.cnt.getContent(idx);
          if (content) {
            this.cnt.rowIdx = idx; // TODO: do this in the nav event?
            GlobalNavEvents.selectContent(content, this.cnt);
          }
        }
      },
    });
  }

  public ngOnDestroy() {
    if (this.sub) {
      this.sub.unsubscribe();
    }
  }

  fullLoadContainer(cnt: Container) {
    console.log('Fully load container from btn click from nav');
    this._contentedService.fullLoadDir(cnt).subscribe({
      next: (loadedDir: Container) => {
        console.log('Fully loaded up the container', loadedDir);
      },
      error: err => {
        GlobalBroadcast.error('Failed to load container', err);
      },
    });
  }

  next() {
    GlobalNavEvents.nextContent(this.cnt);
  }

  nextContainer() {
    GlobalNavEvents.nextContainer();
  }

  prev() {
    GlobalNavEvents.prevContent(this.cnt);
  }

  prevContainer() {
    GlobalNavEvents.prevContainer();
  }
}
