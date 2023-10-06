/*
 * Class for representing the tasks in the background task queue.
 */ 
import * as _ from 'lodash-es';

export class TaskRequest {
    id: string;
    content_id: string;
    created_at: Date|undefined;
    updated_at: Date|undefined;
    status: string;
    operation: string;
    number_of_screens: number;
    start_time_seconds: number;

    codec: string;
    width: number;
    height: number;

    message: string;
    err_msg: string;

    constructor(obj: any) {
        Object.assign(this, obj);

        this.created_at = obj.created_at ? new Date(obj.created_at) : undefined;
        this.updated_at = obj.created_at ? new Date(obj.updated_at) : undefined;
    }
}
