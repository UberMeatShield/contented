import { Observable } from 'rxjs';
import { EventEmitter } from '@angular/core';


export class Message {
    constructor(
        public msg: string,
        public category: string,
        public info: any,
        public channel: string
    ) {

    }
}

export class MessageBroadcast {

    public evts: EventEmitter<Message> =  new EventEmitter<Message>()

    constructor(public channelName: string) {

    }

    evt(msg: string = '', obj: any) {
        console.info(msg, obj);
        this.evts.emit(new Message(msg, 'evt', obj, this.channelName));
    }

    error(msg: string, err: any = {}) {
        console.error(msg, err);
        this.evts.emit(new Message(msg, 'error', err, this.channelName));
    }
}


export const GlobalBroadcast = new MessageBroadcast('global');