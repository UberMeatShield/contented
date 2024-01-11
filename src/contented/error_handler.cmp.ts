import * as _ from 'lodash';
import { Subscription } from 'rxjs';
import {OnInit, OnDestroy, Component, Input, ViewChild, AfterViewInit, Inject} from '@angular/core';
import {MatSnackBar} from '@angular/material/snack-bar';
import {
    MatDialog,
    MatDialogRef,
    MAT_DIALOG_DATA
} from '@angular/material/dialog';

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

    constructor(private snack: MatSnackBar, public dialog: MatDialog) {

    }

    ngOnInit() {
        this.broadcast = this.broadcast || GlobalBroadcast;
        this.sub = this.broadcast.evts.subscribe({
            next: (evt: Message) => {
                if (evt.category === "error") {
                    this.showError(evt);
                }
            }
        });
        setTimeout(() => {
            GlobalBroadcast.error("A bad thing goes bad");
        });
    }

    ngOnDestroy() {
        if (this.sub) {
            this.sub.unsubscribe();
        }
    }

    getErrorCount() {
        return _.filter(this.events, e => e.count > 0).length;
    }

    hasErrors() {
        return !_.isEmpty(_.filter(this.events, e => e.count > 0));
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
        this.snack.open(evt.msg, 'dismiss', {panelClass: 'error', duration: 2000});
    }

    viewErrors() {
        console.log("Open a dialog with a summary of the errors");
        const errors = _.values(this.events);
        const dialogRef = this.dialog.open(ErrorDialogCmp, {
            data: {
                errors: errors,
                width: '90%',
                height: '100%',
                maxWidth: '100vw',
                maxHeight: '100vh',
            }
        });
        dialogRef.afterClosed().subscribe(result => {
            console.log("Closing the dialog", result);
        });
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


// This just doesn't seem like a great approach :(
@Component({
    selector: 'error-dialog',
    templateUrl: 'error_dialog.ng.html'
})
export class ErrorDialogCmp implements AfterViewInit {

    public errors: Array<CountMessages>

    constructor(public dialogRef: MatDialogRef<ErrorDialogCmp>, @Inject(MAT_DIALOG_DATA) public data) {
        this.errors = data.errors;
    }

    dismiss(err: CountMessages) {
        err.count = 0;
        if (_.isEmpty(_.filter(this.errors, err => err.count > 0))) {
            this.dialogRef.close();
        } 
    }

    ngAfterViewInit() {
        console.log("After view init");
    }
}