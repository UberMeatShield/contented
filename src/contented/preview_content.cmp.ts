import { OnInit, Component, Input, HostListener, Output, EventEmitter } from '@angular/core';
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
    selector: 'preview-content-cmp',
    templateUrl: './preview_content.ng.html',
    standalone: false
})
export class PreviewContentCmp {
  @Input() content!: Content;
  @Input() previewWidth: number = 480;
  @Input() previewHeight: number = 480;
  @Input() active: boolean = false;

  @Output() clickEvt: EventEmitter<Content> = new EventEmitter<Content>();

  constructor() {}

  public clickContent(content: Content) {
    this.clickEvt.emit(content);
  }
}
