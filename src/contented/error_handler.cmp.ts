import * as _ from 'lodash';
import {OnInit, OnDestroy, Component, Input, EventEmitter} from '@angular/core';
import { Subscription } from 'rxjs';
import { GlobalBroadcast, MessageBroadcast, Message } from './global_message';

interface CountMessages {
    count: number;
    message: Message;
    uxVisible: boolean;
}

@Component({
    selector: 'error-handler-cmp',
    templateUrl: 'error_handler.ng.html',
})
export class ErrorHandlerCmp implements OnInit, OnDestroy {

    @Input() broadcast: MessageBroadcast;
    public events: {[id: string]: CountMessages} = {};
    public sub: Subscription;

    ngOnInit() {
        this.broadcast = this.broadcast || GlobalBroadcast;
        this.sub = this.broadcast.evts.subscribe({
            next: (evt: Message) => {
                if (evt.category === "error") {
                    this.showError(evt);
                }
            }
        });
    }

    ngOnDestroy() {
        if (this.sub) {
            this.sub.unsubscribe();
        }
    }

    showError(evt: Message) {
        let err = this.events[evt.msg];
        if (!err) {
            err = {
                message: evt,
                count: 1,
                uxVisible: false,
            }
        } else {
            err.count++;
        }
        this.events[evt.msg] = err;
    }

    // Sort, get a count porbably.
    getErrorKeys() {
        return (_.keys(this.events) || []).sort();
    }

    reset() {
        this.events = {};
    }

    clear(msg: string) {
        if (this.events[msg]) {
            delete(this.events[msg]);
        }
    }
}