import {OnInit, Component, EventEmitter, Input, Output, HostListener, ViewChild} from '@angular/core';
import {ContentedService} from './contented_service';
import {Container} from './container';
import {Content} from './content';
import {GlobalNavEvents} from './nav_events';
import {MatRipple} from '@angular/material/core';
import {MatAutocomplete} from '@angular/material/autocomplete';
import {FormControl} from '@angular/forms';
import {Observable} from 'rxjs';
import {map, startWith} from 'rxjs/operators';

import * as _ from 'lodash';
import * as $ from 'jquery';

@Component({
    selector: 'contented-nav',
    templateUrl: 'contented_nav.ng.html'
})
export class ContentedNavCmp implements OnInit {

    @ViewChild(MatRipple) ripple: MatRipple;
    @ViewChild(MatAutocomplete) matAutocomplete: MatAutocomplete;
    @Input() navEvts;
    @Input() loading: boolean;
    @Input() containers: Array<Container>
    @Input() noKeyPress = false;
    @Input() title = '';

    public containerFilter = new FormControl<string>("");
    public filteredContainers: Observable<Container[]>;

    constructor(public _contentedService: ContentedService) {

    }

    ngOnInit() {
        this.navEvts = this.navEvts || GlobalNavEvents.navEvts;
        this.filteredContainers = this.containerFilter.valueChanges.pipe(
            startWith(""),
            map(value => value ? this.filter(value) : this.containers)
        );
    }

    public filter(value: string) {
        let lcVal = value.toLowerCase();
        return _.filter(this.containers, c => {
            return c.name.toLowerCase().includes(lcVal);
        });
    }

    public displaySelection(id: string) {
        let cnt = _.find(this.containers, {id: id});
        return cnt ? cnt.name : "";
    }

    public selectedContainer(cnt: Container) {
        GlobalNavEvents.selectContainer(cnt);

        // If this is not in a delay it will race condition with the selection opening / closing.
        _.delay(() => {
            let filterEl = $("#CONTENT_FILTER");
            filterEl.blur();

            // We want to use the container value setValue to ensure the autocomplete doesn't
            // explode.  Using the dom element itself breaks the dropdown a little bit.
            this.containerFilter.setValue("");
        }, 10);
    }

    // A lot of this stuff is just black magic off stack overflow...
    // it is not obvious it works from the documentation
    public chooseFirstOption() {
        // On enter should turn off focus
        if (this.matAutocomplete.options.first) {
            this.matAutocomplete.options.first.select();
        }
    }

    // On the document keypress events, listen for them (probably need to set them only to component somehow)
    @HostListener('document:keyup', ['$event'])
    public keyPress(evt: KeyboardEvent) {
        // Adds a ripple effect on the buttons (probably should calculate the +32,+20 on element position
        // plus padding etc)  The x,y for a ripple is based on the viewport seemingly.
        if (!/[a-z]/.test(evt.key) && evt.key !== "Escape") {
            // We don't want to freak out and lookup #BTN_} etc.
            return;
        }

        let nodeName = _.get(evt.target, 'nodeName');
        let ignoreNodes = ["TEXTAREA", "INPUT", "SELECT"];
        if (ignoreNodes.includes(nodeName)) {
            return;
        }
        this.handleKey(evt.key);

        let btn = $(`#BTN_${evt.key}`)
        let pos = btn.offset();
        if (pos) {
            // console.log("Position and btn value", pos, btn.val());
            let x = pos.left + 32;
            let y = pos.top + 20;
            let rippleRef = this.ripple.launch(x, y, {
                persistent: true,
                radius: 24,
            });
            _.delay(() => {
                rippleRef.fadeOut();
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
            case 'q':
                GlobalNavEvents.hideFullScreen();
                break;
            case 'Escape':
                // I think it should potentially have a different action
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

