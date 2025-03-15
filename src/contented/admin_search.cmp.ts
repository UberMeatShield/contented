import { SearchCmp } from './search.cmp';
import { Component, Input } from '@angular/core';
import { Content } from './content';
import { GlobalNavEvents } from './nav_events';
import { GlobalBroadcast } from './global_message';
import * as _ from 'lodash';
// TODO: When styling out the search add a hover and hover text to make it
// more obvious when something can be clicked.
@Component({
  selector: 'admin-search-cmp',
  templateUrl: './search.ng.html',
})
export class AdminSearchCmp extends SearchCmp {
  @Input() showToggleDuplicate: boolean = true;

  contentClicked(mc: Content) {
    console.log('Click a content element to open the editor for it.');
    window.open(`/admin_ui/editor_content/${mc.id}`);
  }

  removeDuplicate(mc: Content) {
    if (mc.duplicate) {
      GlobalNavEvents.removeDuplicate(mc);
      this._contentedService.removeContent(mc.id).subscribe({
        next: res => {
          console.log('Removed duplicate', res, mc.id);
          this.content = _.filter(this.content, content => content.id !== mc.id);
        },
        error: err => {
          GlobalBroadcast.error('Error removing duplicate', err);
        },
      });
    } else {
      console.log('Not a duplicate', mc);
    }
  }
}
