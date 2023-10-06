import { Component, OnInit, AfterViewInit, Input, Output, EventEmitter, ViewChild} from '@angular/core';
import {finalize} from 'rxjs/operators';
import {ContentedService} from './contented_service';
import {TaskRequest} from './task_request';

@Component({
  selector: 'task-request-cmp',
  templateUrl: './taskrequest.ng.html',
})
export class TaskRequestCmp implements OnInit {

    @Input() contentID: string = "";
    @Input() reloadEvt: EventEmitter<any>;
    @Input() pageSize = 25;

    public loading = false;
    public tasks: Array<TaskRequest>;

    constructor(public _service: ContentedService) {
    }

    ngOnInit() {
      this.loadTasks(this.contentID);
      if (this.reloadEvt) {
        this.reloadEvt.subscribe(() => {
          this.loadTasks(this.contentID);
        }, console.error)
      }
    }

    loadTasks(contentID: string) {
      this.loading = true;
      this._service.getTasks(this.contentID, 1, this.pageSize).pipe(finalize(() => this.loading = false)).subscribe(
        (tasks) => {
          this.tasks = tasks;
        },
        console.error
      );

    }
}
