import { ComponentFixture, TestBed, fakeAsync, tick } from '@angular/core/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { By } from '@angular/platform-browser';
import { DebugElement } from '@angular/core';

import { ContentedModule } from '../contented/contented_module';
import { TaskRequestCmp } from './taskrequest.cmp';
import { TaskRequest } from './task_request';
import { MockData } from '../test/mock/mock_data';
import { RouterTestingModule } from '@angular/router/testing';

declare var $;

describe('TaskRequestCmp', () => {
  let component: TaskRequestCmp;
  let fixture: ComponentFixture<TaskRequestCmp>;

  let httpMock: HttpTestingController;
  let el: HTMLElement;
  let de: DebugElement;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        NoopAnimationsModule,
        ContentedModule,
        HttpClientTestingModule,
        RouterTestingModule.withRoutes([{ path: 'admin_ui/tasks', component: TaskRequestCmp }]),
      ],
      declarations: [TaskRequestCmp],
    });
    fixture = TestBed.createComponent(TaskRequestCmp);
    component = fixture.componentInstance;

    httpMock = TestBed.inject(HttpTestingController);
    de = fixture.debugElement.query(By.css('.task-request-cmp'));
    el = de.nativeElement;
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('On create we should query for tasks', () => {
    const contentID = 'abc';
    component.contentID = contentID;
    expect(component).toBeTruthy();
    fixture.detectChanges();

    let req = httpMock.expectOne(r => {
      return r.url.includes('/task_requests') && r.params.get('content_id') === contentID;
    });
    req.flush(MockData.taskRequests());
    fixture.detectChanges();

    expect($('.task-cancel-btn').length).toEqual(2);
  });

  it('Should be trying to load tasks', fakeAsync(() => {
    fixture.detectChanges();

    let req = httpMock.expectOne(r => r.url.includes('/task_requests'));
    req.flush(MockData.taskRequests());
    tick(1000);

    expect(component.tasks?.length).withContext('The tasks should be set').toEqual(6);

    fixture.detectChanges();
    expect($('.task-operation').length).withContext('Render the tasks.').toEqual(6);
    tick(1000);
  }));

  it('Can render task information about a duplicate video', () => {
    fixture.detectChanges();

    const tasks = MockData.taskRequests();
    const task = { ...tasks.results[0] };
    task.message = JSON.stringify([
      {
        keep_id: 'db1c539f-e5e2-4e51-a675-85a2ab63cedf',
        container_id: 'dce293d6-7abf-4d4c-9802-3e809e5877a6',
        container_name: 'test_encoding',
        duplicate_id: 'f0cfb1c3-6e7e-4dbf-acfc-7ab2bc9a3e80',
        keep_src: 'SampleVideo_1280x720_1mb_h265.mp4',
        duplicate_src: 'SampleVideo_1280x720_1mb.mp4',
      },
    ]);
    task.operation = 'detect_duplicates';

    const taskCheck = new TaskRequest(task);
    expect(taskCheck.complexMessage).withContext('It should detect a complex return').toBeDefined();

    let req = httpMock.expectOne(r => r.url.includes('/task_requests'));
    req.flush({ total: 1, results: [task] });
    fixture.detectChanges();
    expect($('.task-operation').length).withContext('Render the tasks.').toEqual(1);
    expect($('.duplicate-link').length).withContext('Duplicate tasks should provide links').toEqual(1);
  });
});
