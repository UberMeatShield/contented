import {OnInit, Component, EventEmitter, Input, Output, HostListener, ViewChild} from '@angular/core';
import {ContentedService} from './contented_service';
import {Container} from './container';
import {Media} from './media';
import {GlobalNavEvents} from './nav_events';
import {MatRipple} from '@angular/material/core';
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
    @Input() navEvts;
    @Input() loading: boolean;
    @Input() containers: Array<Container>

    public containerFilter = new FormControl('');
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

    // On the document keypress events, listen for them (probably need to set them only to component somehow)
    @HostListener('document:keypress', ['$event'])
    public keyPress(evt: KeyboardEvent) {

        // Adds a ripple effect on the buttons (probably should calculate the +32,+20 on element position
        // plus padding etc)  The x,y for a ripple is based on the viewport seemingly.
        let btn = $(`#BTN_${evt.key}`)
        let pos = btn.offset();
        if (pos) {
            console.log("Position and btn value", pos, btn.val());
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
        this.handleKey(evt.key);
    }

    public handleKey(key: string) {
        console.log("Handle keypress", key);
        switch (key) {
            case 'w':
                GlobalNavEvents.prevContainer();
                break;
            case 's':
                GlobalNavEvents.nextContainer();
                break;
            case 'a':
                GlobalNavEvents.prevMedia();
                break;
            case 'd':
                GlobalNavEvents.nextMedia();
                break;
            case 'e':
                GlobalNavEvents.viewFullScreen();
                break;
            case 'q':
                GlobalNavEvents.hideFullScreen();
                break;
            case 'f':
                GlobalNavEvents.loadMoreMedia();
                break;
            case 'x':
                GlobalNavEvents.saveMedia();
                break;
            default:
                break;
        }

    }
}

