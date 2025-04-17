import { OnInit, Component, Input, HostListener, ViewChild, EventEmitter } from '@angular/core';
import { ContentedService } from './contented_service';
import { Container, getFavorites } from './container';
import { GlobalNavEvents, NavEventMessage } from './nav_events';
import { MatRipple } from '@angular/material/core';
import { MatAutocomplete } from '@angular/material/autocomplete';
import { FormControl } from '@angular/forms';
import type { Observable } from 'rxjs';
import { map, startWith } from 'rxjs/operators';

import _ from 'lodash';

@Component({
    selector: 'contented-nav',
    templateUrl: 'contented_nav.ng.html',
    standalone: false
})
export class ContentedNavCmp implements OnInit {
  @ViewChild(MatRipple) ripple: MatRipple | undefined;
  @ViewChild(MatAutocomplete) matAutocomplete: MatAutocomplete | undefined;
  @Input() navEvts: EventEmitter<NavEventMessage> | undefined;
  @Input() loading: boolean = false;
  @Input() containers: Array<Container> = [];
  @Input() noKeyPress = false;
  @Input() title = '';
  @Input() showFavorites = true;

  public containerFilter = new FormControl<string | null>('');
  public filteredContainers: Observable<Container[]>;
  public favoriteContainer: Container | undefined;

  constructor(public _contentedService: ContentedService) {
    this.favoriteContainer = getFavorites();
    this.filteredContainers = this.containerFilter.valueChanges.pipe(
      startWith(''),
      map(value => (value ? this.filter(value) : this.containers))
    );
  }

  ngOnInit() {
    this.navEvts = this.navEvts || GlobalNavEvents.navEvts;
  }

  public toggleFavorites() {
    console.log('Toggle favorites');
    GlobalNavEvents.toggleFavoriteVisibility();
  }

  public filter(value: string) {
    const lcVal = value.toLowerCase();
    return _.filter(this.containers, c => {
      return c.name.toLowerCase().includes(lcVal);
    });
  }

  public displaySelection(id: number) {
    const cnt = _.find(this.containers, { id: id });
    return cnt ? cnt.name : '';
  }

  public selectedContainer(cnt: Container) {
    GlobalNavEvents.selectContainer(cnt);

    // If this is not in a delay it will race condition with the selection opening / closing.
    _.delay(() => {
      const filterEl = document.getElementById('CONTENT_FILTER');
      filterEl?.blur();

      // We want to use the container value setValue to ensure the autocomplete doesn't
      // explode.  Using the dom element itself breaks the dropdown a little bit.
      this.containerFilter.setValue('');
    }, 10);
  }

  // A lot of this stuff is just black magic off stack overflow...
  // it is not obvious it works from the documentation
  public chooseFirstOption() {
    // On enter should turn off focus
    if (this.matAutocomplete?.options.first) {
      this.matAutocomplete.options.first.select();
    }
  }

  getOffset(element: HTMLElement) {
    if (!element.getClientRects().length) {
      return { top: 0, left: 0 };
    }

    let rect = element.getBoundingClientRect();
    let win = element.ownerDocument.defaultView;
    return {
      top: rect.top + window.pageYOffset,
      left: rect.left + window.pageXOffset,
    };
  }

  // On the document keypress events, listen for them (probably need to set them only to component somehow)
  @HostListener('document:keyup', ['$event'])
  public keyPress(evt: KeyboardEvent) {
    // Adds a ripple effect on the buttons (probably should calculate the +32,+20 on element position
    // plus padding etc)  The x,y for a ripple is based on the viewport seemingly.
    if (!/[a-z]/.test(evt.key) && evt.key !== 'Escape') {
      // We don't want to freak out and lookup #BTN_} etc.
      return;
    }

    let nodeName = _.get(evt.target, 'nodeName');
    let ignoreNodes = ['TEXTAREA', 'INPUT', 'SELECT'];
    if (nodeName && ignoreNodes.includes(nodeName)) {
      return;
    }
    this.handleKey(evt.key);

    let btn = document.getElementById(`BTN_${evt.key}`);

    let pos = btn ? this.getOffset(btn) : { top: 0, left: 0 };
    if (pos) {
      // console.log("Position and btn value", pos, btn.val());
      let x = pos.left + 32;
      let y = pos.top + 16;
      let rippleRef = this.ripple?.launch(x, y, {
        persistent: true,
        radius: 12,
      });
      _.delay(() => {
        rippleRef?.fadeOut();
      }, 250);
    }
  }

  public handleKey(key: string) {
    // console.log("Handle keypress", key);
    switch (key) {
      case 'w':
        GlobalNavEvents.prevContainer();
        break;
      case 's':
        GlobalNavEvents.nextContainer();
        break;
      case 'a':
        GlobalNavEvents.prevContent();
        break;
      case 'd':
        GlobalNavEvents.nextContent();
        break;
      case 'e':
        GlobalNavEvents.viewFullScreen();
        break;
      case 't':
        GlobalNavEvents.toggleFavorite();
        break;
      case 'q':
        GlobalNavEvents.hideFullScreen();
        break;
      case 'Escape':
        // I think it should potentially have a different action for escape vs q
        GlobalNavEvents.hideFullScreen();
        break;
      case 'f':
        GlobalNavEvents.loadMoreContent();
        break;
      case 'x':
        GlobalNavEvents.saveContent();
        break;
      default:
        break;
    }
  }
}
