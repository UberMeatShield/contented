import { Component, OnInit, AfterViewInit, Input, Output, EventEmitter, ViewChild } from '@angular/core';
import { finalize, debounceTime, distinctUntilChanged } from 'rxjs/operators';
import { ContentedService } from './contented_service';
import { TaskRequest, TASK_STATES } from './task_request';
import { MatTableDataSource } from '@angular/material/table';
import { FormBuilder, NgForm, FormControl, FormGroup } from '@angular/forms';

import * as _ from 'lodash';

@Component({
    selector: 'tasks-cmp',
    templateUrl: './tasks.ng.html',
    standalone: false
})
export class TasksCmp implements OnInit {
  constructor(
    public _service: ContentedService,
    fb: FormBuilder
  ) {}
  ngOnInit() {
    console.log('Attempt to load up tasks and build a graph');
  }
}
