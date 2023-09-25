import {Subscription} from 'rxjs';
import {OnInit, OnDestroy, Component, EventEmitter, Input, Output, HostListener} from '@angular/core';
import {ContentedService} from './contented_service';
import {Container, LoadStates} from './container';
import {Content} from './content';
import {finalize, switchMap} from 'rxjs/operators';

import {ActivatedRoute, Router, ParamMap} from '@angular/router';
import {GlobalNavEvents, NavTypes} from './nav_events';
import {FormGroup, FormBuilder, FormControl, Validators} from '@angular/forms';

import * as _ from 'lodash';

@Component({
    selector: 'container-nav',
    templateUrl: 'container_nav.ng.html'
})
export class ContainerNavCmp implements OnInit, OnDestroy {

    // This is actually required
    @Input() cnt: Container;

    // Do we actually care?
    @Input() totalContainers: number = 0;

    // current view Item should be something you trigger per directory (move view ?)
    public currentContent: Content;
    public ContainerLoadStates = LoadStates;

    // idx and current view item might be better as a top level nav / hover should be allowed?
    @Input() active: boolean = false; // Is our container active

    // rowIdx should be independently controlled for each directory
    @Output() navEvt: EventEmitter<any> = new EventEmitter<any>();
    @Input() rowIdx: number = 0; // Which content item is selected
    @Input() idx: number = 0; // What is our index compared to other containers

    private sub: Subscription;

    public navForm: FormGroup;
    public idxControl: FormControl<number>;

    constructor(public fb: FormBuilder, public _contentedService: ContentedService) {
    }

    public ngOnInit() {
        this.idxControl = new FormControl(this.idx || this.cnt.rowIdx || 0, Validators.required);
        this.navForm = this.fb.group({
            idxControl: this.idxControl
        });

        this.sub = GlobalNavEvents.navEvts.subscribe(evt => {
            if (evt.action == NavTypes.SELECT_MEDIA && evt.cnt == this.cnt && evt.content) {
                //console.log("Container Nav found select content", evt, evt.cnt.name);
                this.currentContent = evt.content;
                this.idxControl.setValue(this.cnt.rowIdx);
            }
        });
        // The select event can trigger BEFORE a render loop so on a new render
        // ensure we at least get our current content (should be correct given the rowIdx)
        if (this.cnt) {
            this.currentContent = this.cnt.getContent();
        }

        this.navForm.get("idxControl").valueChanges.subscribe(
            idx => {
                if (idx != this.cnt.rowIdx) {
                    let content = this.cnt.getContent(idx);
                    if (content) {
                        console.log("Index changed via input", idx);
                        this.cnt.rowIdx = idx;  // TODO: do this in the nav event?
                        GlobalNavEvents.selectContent(content, this.cnt);
                    }
                }
            }
        );
    }

    public ngOnDestroy() {
        if (this.sub) {
            this.sub.unsubscribe();
        }
    }

    fullLoadContainer(cnt: Container) {
        console.log("Fully load container from btn click from nav");
        this._contentedService.fullLoadDir(cnt).subscribe(
            (loadedDir: Container) => {
                console.log("Fully loaded up the container", loadedDir);
            },
            err => {
                console.error("Failed to load", err);
             }
        );
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

