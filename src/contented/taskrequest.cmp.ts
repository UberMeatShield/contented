import { Component, OnInit, AfterViewInit, Input, Output, EventEmitter, ViewChild} from '@angular/core';
import {finalize} from 'rxjs/operators';
import {ContentedService} from './contented_service';
import {TaskRequest} from './task_request';
import { MatTableDataSource} from '@angular/material/table';
import {ActivatedRoute, Router, ParamMap} from '@angular/router';

@Component({
  selector: 'task-request-cmp',
  templateUrl: './taskrequest.ng.html',
})
export class TaskRequestCmp implements OnInit {

    @Input() contentID: string = "";
    @Input() pageSize = 25;
    @Input() reloadEvt: EventEmitter<any>; // Do you want to reload the task queue
    @Output() taskUpdated: EventEmitter<TaskRequest> = new EventEmitter<TaskRequest>;

    public loading = false;
    public tasks: Array<TaskRequest>;

    displayedColumns: string[] = ['operation', 'status', 'created_at', 'updated_at', 'message', 'created_id', 'error'];
    dataSource = new MatTableDataSource<TaskRequest>([]);

    constructor(public _service: ContentedService, public route: ActivatedRoute) {

    }

    ngOnInit() {
      this.route.paramMap.pipe().subscribe(
        evt => {
          // Currently it will reload because content ID information is updated in the editor pane.
           console.log("Route event", evt); 
        }
      )

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
          this.dataSource = new MatTableDataSource<TaskRequest>(tasks || [])
        },
        console.error
      );
    }
    // Enable a polling method that will check for when a task is done (in editor or here?)
}
