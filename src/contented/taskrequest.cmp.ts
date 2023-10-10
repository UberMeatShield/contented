import { Component, OnInit, AfterViewInit, Input, Output, EventEmitter, ViewChild} from '@angular/core';
import {finalize} from 'rxjs/operators';
import {ContentedService} from './contented_service';
import {TaskRequest} from './task_request';
import { MatTableDataSource} from '@angular/material/table';
import {ActivatedRoute, Router, ParamMap} from '@angular/router';

import * as _ from 'lodash-es';

@Component({
  selector: 'task-request-cmp',
  templateUrl: './taskrequest.ng.html',
})
export class TaskRequestCmp implements OnInit {

    @Input() contentID: string = "";
    @Input() pageSize = 25;
    @Input() reloadEvt: EventEmitter<any>; // Do you want to reload the task queue
    @Output() taskUpdated: EventEmitter<TaskRequest> = new EventEmitter<TaskRequest>;
    @Input() checkStates = false;

    public loading = false;
    public tasks: Array<TaskRequest>;

    displayedColumns: string[] = ['operation', 'status', 'created_at', 'updated_at', 'message', 'created_id', 'error'];
    dataSource = new MatTableDataSource<TaskRequest>([]);

    //constructor(public _service: ContentedService, public route: ActivatedRoute) {
    constructor(public _service: ContentedService) {
    }

    ngOnInit() {
      this.loadTasks(this.contentID);


      if (this.reloadEvt) {
        this.reloadEvt.subscribe(() => {
          // Reload and consider we should have a watcher

          _.delay(() => {
            this.loadTasks(this.contentID);
          }, 2000);
        }, console.error)
      }
      if (this.checkStates) {
        this.pollStart();
      }
    }

    loadTasks(contentID: string, watching: Array<TaskRequest> = []) {
      this.loading = true;
      this._service.getTasks(this.contentID, 1, this.pageSize).pipe(finalize(() => this.loading = false)).subscribe(
        (tasks) => {
          this.tasks = tasks;
          this.dataSource = new MatTableDataSource<TaskRequest>(tasks || [])
        },
        console.error
      );
    }

    checkComplete(tasks: Array<TaskRequest>, watching: Array<TaskRequest> = []) {
      let check = _.keyBy(watching, 'id');
      (tasks || []).forEach(task => {
        if (check[task.id] && task.isComplete()) {
          console.log("Watching tasks")
          this.taskUpdated.emit(task);
        }
      });
    }

    pollStart() {
      this.pollTasks();
      _.delay(() => {
        this.pollStart()
      }, 1000 * 10);
    }

    pollTasks() {
      let notComplete: Array<TaskRequest> = _.filter(this.tasks, task => {
        return task.isComplete();
      }) || [];
      this.loadTasks(this.contentID, notComplete);
    }
}
